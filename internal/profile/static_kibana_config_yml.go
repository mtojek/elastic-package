// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package profile

import "path/filepath"

const kibanaConfigYml = `  
server.name: kibana
server.host: "0.0.0.0"

elasticsearch.hosts: [ "http://elasticsearch:9200" ]
elasticsearch.username: elastic
elasticsearch.password: changeme

xpack.monitoring.ui.container.elasticsearch.enabled: true

xpack.fleet.enabled: true
xpack.fleet.registryUrl: "http://package-registry:8080"
xpack.fleet.agents.enabled: true
xpack.fleet.agents.elasticsearch.host: "http://elasticsearch:9200"
xpack.fleet.agents.fleet_server.hosts: ["http://fleet-server:8220"]

xpack.encryptedSavedObjects.encryptionKey: "12345678901234567890123456789012"
`

// KibanaConfigFile is the main kibana config file
const KibanaConfigFile configFile = "kibana.config.yml"

// newKibanaConfig returns a Managed Config
func newKibanaConfig(_ string, profilePath string) (*simpleFile, error) {
	return &simpleFile{
		name: string(KibanaConfigFile),
		path: filepath.Join(profilePath, profileStackPath, string(KibanaConfigFile)),
		body: kibanaConfigYml,
	}, nil
}
