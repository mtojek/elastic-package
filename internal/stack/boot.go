// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package stack

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/elastic/elastic-package/internal/builder"
	"github.com/elastic/elastic-package/internal/files"
	"github.com/elastic/elastic-package/internal/install"
)

// DockerComposeProjectName is the name of the Docker Compose project used to boot up
// Elastic Stack containers.
const DockerComposeProjectName = "elastic-package-stack"

// BootUp method boots up the testing stack.
func BootUp(options Options) error {
	buildPackagesPath, found, err := builder.FindBuildPackagesDirectory()
	if err != nil {
		return errors.Wrap(err, "finding build packages directory failed")
	}

	stackPackagesDir, err := install.StackPackagesDir()
	if err != nil {
		return errors.Wrap(err, "locating stack packages directory failed")
	}

	err = files.ClearDir(stackPackagesDir)
	if err != nil {
		return errors.Wrap(err, "clearing package contents failed")
	}

	if found {
		fmt.Printf("Custom build packages directory found: %s\n", buildPackagesPath)
		err = files.CopyAll(buildPackagesPath, stackPackagesDir)
		if err != nil {
			return errors.Wrap(err, "copying package contents failed")
		}
	}

	err = dockerComposeBuild(options)
	if err != nil {
		return errors.Wrap(err, "building docker images failed")
	}

	toBeStopped := determineServicesToBeStopped(options.Services)
	if len(toBeStopped) > 0 {
		err = dockerComposeDown(Options{
			Services: toBeStopped,
		})
		if err != nil {
			return errors.Wrap(err, "stopping docker containers failed")
		}
	}

	err = dockerComposeUp(options.WithServices(options.Services))
	if err != nil {
		return errors.Wrap(err, "running docker-compose failed")
	}
	return nil
}

func determineServicesToBeStopped(toBeStarted []string) []string {
	toBeStopped := map[string]bool{
		"elasticsearch":    true,
		"kibana":           true,
		"package-registry": true,
		"elastic-agent":    true,
	}

	for _, service := range toBeStarted {
		if _, ok := toBeStopped[service]; ok {
			toBeStopped[service] = false
		}
	}

	var t []string
	for service, isUp := range toBeStopped {
		if !isUp {
			continue
		}
		t = append(t, service)
	}
	return t
}

// TearDown method takes down the testing stack.
func TearDown(options Options) error {
	err := dockerComposeDown(options)
	if err != nil {
		return errors.Wrap(err, "stopping docker containers failed")
	}
	return nil
}
