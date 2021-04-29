// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package archetype

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/elastic/elastic-package/internal/packages"
)

func TestDataStream(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		pd := createPackageDescriptorForTest()
		dd := createDataStreamDescriptorForTest()

		err := createAndCheckDataStream(t, pd, dd)
		require.NoError(t, err)
	})
}

func createDataStreamDescriptorForTest() DataStreamDescriptor {
	return DataStreamDescriptor{
		Manifest: packages.DataStreamManifest{
			Name:  "go_unit_test_data_stream",
			Title: "Go Unit Test Data Stream",
			Type:  "logs",
		},
	}
}

func createAndCheckDataStream(t require.TestingT, pd PackageDescriptor, dd DataStreamDescriptor) error {
	wd, err := os.Getwd()
	require.NoError(t, err)

	tempDir, err := ioutil.TempDir("", "archetype-create-data-stream-")
	require.NoError(t, err)

	os.Chdir(tempDir)
	defer func() {
		os.Chdir(wd)
		os.RemoveAll(tempDir)
	}()

	err = CreatePackage(pd)
	require.NoError(t, err)

	packageRoot := filepath.Join(tempDir, pd.Manifest.Name)
	dd.PackageRoot = packageRoot

	err = CreateDataStream(dd)
	require.NoError(t, err)

	err = checkPackage(pd.Manifest.Name)
	return err
}
