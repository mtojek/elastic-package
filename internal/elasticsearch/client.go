// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package elasticsearch

import (
	"os"

	"github.com/pkg/errors"

	"github.com/elastic/go-elasticsearch/v7"

	"github.com/elastic/elastic-package/internal/stack"
)

// Client method creates new instance of the Elasticsearch client.
func Client() (*elasticsearch.Client, error) {
	host := os.Getenv(stack.ElasticsearchHostEnv)
	if host == "" {
		return nil, stack.UndefinedEnvError(stack.ElasticsearchHostEnv)
	}

	username := os.Getenv(stack.ElasticsearchUsernameEnv)
	password := os.Getenv(stack.ElasticsearchPasswordEnv)

	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{host},
		Username:  username,
		Password:  password,
	})
	if err != nil {
		return nil, errors.Wrap(err, "can't create instance")
	}
	return client, nil
}
