// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package packages

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/elastic/go-ucfg"
	"github.com/elastic/go-ucfg/yaml"
	"github.com/pkg/errors"
)

const (
	// PackageManifestFile is the name of the package's main manifest file.
	PackageManifestFile = "manifest.yml"

	// DataStreamManifestFile is the name of the data stream's manifest file.
	DataStreamManifestFile = "manifest.yml"
)

// VarValue represents a variable value as defined in a package or data stream
// manifest file.
type VarValue struct {
	scalar string
	list   []string
}

// Unpack knows how to parse a variable value from a package or data stream
// manifest file into a VarValue.
func (vv *VarValue) Unpack(cfg *ucfg.Config) error {
	if cfg.IsArray() {
		return cfg.Unpack(&vv.list)
	}
	if !cfg.IsDict() { // is scalar as dict is not supported
		return cfg.Unpack(&vv.scalar)
	}
	return errors.New("unknown variable value")
}

// MarshalJSON knows how to serialize a VarValue into the appropriate
// JSON data type and value.
func (vv VarValue) MarshalJSON() ([]byte, error) {
	if vv.scalar != "" {
		return json.Marshal(vv.scalar)
	} else if vv.list != nil {
		return json.Marshal(vv.list)
	}
	return nil, nil
}

// Variable is an instance of configuration variable (named, typed).
type Variable struct {
	Name    string   `config:"name" json:"name" yaml:"name"`
	Type    string   `config:"type" json:"type" yaml:"type"`
	Default VarValue `config:"default" json:"default" yaml:"default"`
}

// Input is a single input configuration.
type Input struct {
	Type string     `config:"type" json:"type" yaml:"type"`
	Vars []Variable `config:"vars" json:"vars" yaml:"vars"`
}

// PolicyTemplate is a configuration of inputs responsible for collecting log or metric data.
type PolicyTemplate struct {
	Inputs []Input `config:"inputs" json:"inputs" yaml:"inputs"`
}

// PackageManifest represents the basic structure of a package's manifest
type PackageManifest struct {
	Name            string           `config:"name" json:"name" yaml:"name"`
	Title           string           `config:"title" json:"title" yaml:"title"`
	Type            string           `config:"type" json:"type" yaml:"type"`
	Version         string           `config:"version" json:"version" yaml:"version"`
	PolicyTemplates []PolicyTemplate `config:"policy_templates" json:"policy_templates" yaml:"policy_templates"`
}

// DataStreamManifest represents the structure of a data stream's manifest
type DataStreamManifest struct {
	Name          string `config:"name" json:"name" yaml:"name"`
	Title         string `config:"title" json:"title" yaml:"title"`
	Type          string `config:"type" json:"type" yaml:"type"`
	Elasticsearch *struct {
		IngestPipeline *struct {
			Name string `config:"name" json:"name" yaml:"name"`
		} `config:"ingest_pipeline" json:"ingest_pipeline" yaml:"ingest_pipeline"`
	} `config:"elasticsearch" json:"elasticsearch" yaml:"elasticsearch"`
	Streams []struct {
		Input string     `config:"input" json:"input" yaml:"input"`
		Vars  []Variable `config:"vars" json:"vars" yaml:"vars"`
	} `config:"streams" json:"streams" yaml:"streams"`
}

// FindPackageRoot finds and returns the path to the root folder of a package.
func FindPackageRoot() (string, bool, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return "", false, errors.Wrap(err, "locating working directory failed")
	}

	dir := workDir
	for dir != "." {
		path := filepath.Join(dir, PackageManifestFile)
		fileInfo, err := os.Stat(path)
		if err == nil && !fileInfo.IsDir() {
			ok, err := isPackageManifest(path)
			if err != nil {
				return "", false, errors.Wrapf(err, "verifying manifest file failed (path: %s)", path)
			}
			if ok {
				return dir, true, nil
			}
		}

		if dir == "/" {
			break
		}
		dir = filepath.Dir(dir)
	}
	return "", false, nil
}

// FindDataStreamRootForPath finds and returns the path to the root folder of a data stream.
func FindDataStreamRootForPath(workDir string) (string, bool, error) {
	dir := workDir
	for dir != "." {
		path := filepath.Join(dir, DataStreamManifestFile)
		fileInfo, err := os.Stat(path)
		if err == nil && !fileInfo.IsDir() {
			ok, err := isDataStreamManifest(path)
			if err != nil {
				return "", false, errors.Wrapf(err, "verifying manifest file failed (path: %s)", path)
			}
			if ok {
				return dir, true, nil
			}
		}

		if dir == "/" {
			break
		}
		dir = filepath.Dir(dir)
	}
	return "", false, nil
}

// ReadPackageManifest reads and parses the given package manifest file.
func ReadPackageManifest(path string) (*PackageManifest, error) {
	cfg, err := yaml.NewConfigWithFile(path, ucfg.PathSep("."))
	if err != nil {
		return nil, errors.Wrapf(err, "reading file failed (path: %s)", path)
	}

	var m PackageManifest
	err = cfg.Unpack(&m)
	if err != nil {
		return nil, errors.Wrapf(err, "unpacking package manifest failed (path: %s)", path)
	}
	return &m, nil
}

// ReadDataStreamManifest reads and parses the given data stream manifest file.
func ReadDataStreamManifest(path string) (*DataStreamManifest, error) {
	cfg, err := yaml.NewConfigWithFile(path, ucfg.PathSep("."))
	if err != nil {
		return nil, errors.Wrapf(err, "reading file failed (path: %s)", path)
	}

	var m DataStreamManifest
	err = cfg.Unpack(&m)
	if err != nil {
		return nil, errors.Wrapf(err, "unpacking data stream manifest failed (path: %s)", path)
	}

	m.Name = filepath.Base(filepath.Dir(path))
	return &m, nil
}

// FindInputByType returns the input for the provided type.
func (pt *PolicyTemplate) FindInputByType(inputType string) *Input {
	for _, input := range pt.Inputs {
		if input.Type == inputType {
			return &input
		}
	}
	return nil
}

func isPackageManifest(path string) (bool, error) {
	m, err := ReadPackageManifest(path)
	if err != nil {
		return false, errors.Wrapf(err, "reading package manifest failed (path: %s)", path)
	}
	return m.Type == "integration" && m.Version != "", nil // TODO add support for other package types
}

func isDataStreamManifest(path string) (bool, error) {
	m, err := ReadDataStreamManifest(path)
	if err != nil {
		return false, errors.Wrapf(err, "reading package manifest failed (path: %s)", path)
	}
	return m.Title != "" && (m.Type == "logs" || m.Type == "metrics"), nil
}
