// Package openapi provides access to the the most recent version of the OpenAPI
// specification.
//
// It provides methods for loading files of version 2 and up, and converting them
// to v3.
//
// XXX: actually define v3.
// XXX: actually convert v2 to v3
package openapi

import (
	"io/ioutil"

	"github.com/go-yaml/yaml"
	"github.com/jbowes/oag/openapi/v2"
)

// LoadFile loads the OpenAPI file at the given path, returning the document
// definition.
func LoadFile(path string) (*v2.Document, error) {
	d, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var doc v2.Document
	err = yaml.Unmarshal(d, &doc)
	if err != nil {
		return nil, err
	}

	return &doc, err
}
