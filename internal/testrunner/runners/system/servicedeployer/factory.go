// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package servicedeployer

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

const devDeployDir = "_dev/deploy"

// FactoryOptions defines options used to create an instance of a service deployer.
type FactoryOptions struct {
	PackageRootPath    string
	DataStreamRootPath string
}

// Factory chooses the appropriate service runner for the given data stream, depending
// on service configuration files defined in the package or data stream.
func Factory(options FactoryOptions) (ServiceDeployer, error) {
	devDeployPath, err := findDevDeployPath(options)
	if err != nil {
		return nil, errors.Wrapf(err, "can't find \"%s\" directory", devDeployDir)
	}

	serviceDeployerName, err := findServiceDeployer(devDeployDir)
	if err != nil {
		return nil, errors.Wrap(err, "can't find any valid service deployer")
	}

	switch serviceDeployerName {
	case "docker":
		dockerComposeYMLPath := filepath.Join(devDeployPath, serviceDeployerName, "docker-compose.yml")
		if _, err := os.Stat(dockerComposeYMLPath); err == nil {
			return NewDockerComposeServiceDeployer(dockerComposeYMLPath)
		}
	case "tf":
		terraformDirPath := filepath.Join(devDeployPath, serviceDeployerName)
		if _, err := os.Stat(terraformDirPath); err == nil {
			return NewTerraformServiceDeployer(terraformDirPath)
		}
	}
	return nil, fmt.Errorf("unsupported service deployer (name: %s)", serviceDeployerName)
}

func findServiceDeployer(devDeployPath string) (string, error) {
	fis, err := ioutil.ReadDir(devDeployPath)
	if err != nil {
		return "", errors.Wrapf(err, "can't read directory (path: %s)", devDeployDir)
	}

	if len(fis) != 1 {
		return "", fmt.Errorf("expected to find only one service deployer in \"%s\"", devDeployPath)
	}

	deployerFileInfo := fis[0]
	if !deployerFileInfo.IsDir() {
		return "", fmt.Errorf("\"%s\" is expected to be a folder in \"%s\"", deployerFileInfo, devDeployPath)
	}

	return deployerFileInfo.Name(), nil
}

func findDevDeployPath(options FactoryOptions) (string, error) {
	dataStreamDevDeployPath := filepath.Join(options.DataStreamRootPath, devDeployDir)
	_, err := os.Stat(dataStreamDevDeployPath)
	if err == nil {
		return dataStreamDevDeployPath, nil
	} else if !os.IsNotExist(err) {
		return "", errors.Wrapf(err, "stat failed for data stream (path: %s)", dataStreamDevDeployPath)
	}

	packageDevDeployPath := filepath.Join(options.PackageRootPath, devDeployDir)
	_, err = os.Stat(packageDevDeployPath)
	if err == nil {
		return packageDevDeployPath, nil
	} else if !os.IsNotExist(err) {
		return "", errors.Wrapf(err, "stat failed for package (path: %s)", packageDevDeployPath)
	}
	return "", fmt.Errorf("\"%s\" directory doesn't exist", devDeployDir)
}
