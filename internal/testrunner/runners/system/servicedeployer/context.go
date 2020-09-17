// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package servicedeployer

// ServiceContext encapsulates context that is both available to a ServiceDeployer and
// populated by a DeployedService. The fields in ServiceContext may be used in handlebars
// templates in system test configuration files, for example: {{ Hostname }}.
type ServiceContext struct {
	// Name is the name of the service.
	Name string

	// Hostname is the host name of the service, as addressable from
	// the Agent container.
	Hostname string

	// Ports is a list of ports that the service listens on, as addressable
	// from the Agent container.
	Ports []int

	// Local contains the folder path where log files produced by
	// the service are stored on the local filesystem, i.e. where
	// elastic-package is running.
	LogsFolderLocal string
}
