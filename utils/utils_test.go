package utils

import (
	"errors"
	"github.com/stretchr/testify/assert"
	serviceDiscovery "serviceDiscovery/consul"
	"testing"
)

func TestIsDiscoveryServiceEndpoint(t *testing.T) {
	type isDiscoveryServiceEndpointList struct {
		input   string
		output  bool
		message string
	}
	dataItems := []isDiscoveryServiceEndpointList{
		{
			input:   "consul:",
			output:  true,
			message: "only consul keyword",
		},
		{
			input:   "consul:",
			output:  true,
			message: "only consul keyword",
		},
		{
			input:   "consul:[dc1,dc2].dev.serviceA.[tag1,tag2]",
			output:  true,
			message: "valid",
		},
		{
			input:   "",
			output:  false,
			message: "empty",
		},
		{
			input:   "consul",
			output:  false,
			message: "without :",
		},
	}

	for i, item := range dataItems {
		result := isDiscoveryServiceEndpoint(item.input, consul)
		assert.Equal(t, item.output, result, item.message, i)
	}
}

func TestParseSyntax(t *testing.T) {
	type parseListTestItem struct {
		input   string
		result  *serviceDiscovery.QueryString
		err     error
		message string
	}
	dataItems := []parseListTestItem{
		{
			input: "consul:[dc1,dc2].dev.serviceA.[tag1,tag2]",
			result: &serviceDiscovery.QueryString{
				Datacenters: []string{"dc1", "dc2"},
				ServiceName: "serviceA",
				Namespace:   "dev",
				Tags:        []string{"tag1", "tag2"},
			},
			err:     nil,
			message: "simple scenario with namespace",
		},
		{
			input: "consul:[dc 1,dc 2].service A.[tag1,tag2]",
			result: &serviceDiscovery.QueryString{
				Datacenters: []string{"dc 1", "dc 2"},
				ServiceName: "service A",
				Namespace:   "",
				Tags:        []string{"tag1", "tag2"},
			},
			err:     nil,
			message: "simple scenario without namespace",
		},
		{
			input: "consul:[].prod.serviceA.[*]",
			result: &serviceDiscovery.QueryString{
				Datacenters: []string{""},
				ServiceName: "serviceA",
				Namespace:   "prod",
				Tags:        []string{""},
			},
			err:     nil,
			message: "empty dcs and tags",
		},
		{
			input:   "consul[].prod.serviceA.[*]",
			err:     errors.New("bad query syntax"),
			message: "empty dcs and tags",
		},
	}
	for i, item := range dataItems {
		result, err := parseSyntax(item.input)
		assert.Equal(t, item.result, result, item.message, i)
		assert.Equal(t, item.err, err, item.message)
	}
}

func TestParseList(t *testing.T) {
	type parseListTestItem struct {
		inputString string
		resultList  []string
		message     string
	}
	dataItems := []parseListTestItem{
		{
			inputString: "[dc1,dc2,aws-us-central-1]",
			resultList:  []string{"dc1", "dc2", "aws-us-central-1"},
			message:     "Simple scenario with 3dcs",
		},
		{
			inputString: "[]",
			resultList:  []string{""},
			message:     "Empty list :(all)",
		},
		{
			inputString: "[*]",
			resultList:  []string{""},
			message:     "Empty list with * :(all)",
		},
		{
			inputString: "[abc]",
			resultList:  []string{"abc"},
			message:     "List with one dc",
		},
	}
	for _, item := range dataItems {
		result := parseList(item.inputString)
		assert.Equal(t, item.resultList, result, item.message)
	}
}
