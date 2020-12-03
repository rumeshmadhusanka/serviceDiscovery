package utils

import (
	"errors"
	serviceDiscovery "serviceDiscovery/consul"
	"strings"
)

const (
	consul string = "consul:"
)

func isDiscoveryServiceEndpoint(str string, discoveryServiceName string) bool {
	return strings.HasPrefix(str, discoveryServiceName)
}

func parseList(str string) []string {
	s := strings.Split(str, ",")
	for i := range s {
		s[i] = strings.ReplaceAll(s[i], "[", "")
		s[i] = strings.ReplaceAll(s[i], "]", "")
		if s[i] == "*" {
			s[i] = ""
		}
	}
	return s
}

func parseSyntax(value string) (*serviceDiscovery.QueryString, error) {
	//[dc1,dc2].namespace.serviceA.[tag1,tag2]
	split := strings.Split(value, ":")
	if len(split) != 2 {
		return nil, errors.New("bad query syntax")
	}
	str := strings.Split(split[1], ".")
	qCategory := len(str)
	if qCategory == 3 { //datacenters, service name, tags
		queryString := serviceDiscovery.QueryString{
			Datacenters: parseList(str[0]),
			ServiceName: str[1],
			Namespace:   "",
			Tags:        parseList(str[2]),
		}
		return &queryString, nil
	} else if qCategory == 4 { //datacenters, namespace, service name, tags
		queryString := serviceDiscovery.QueryString{
			Datacenters: parseList(str[0]),
			ServiceName: str[2],
			Namespace:   str[1],
			Tags:        parseList(str[3]),
		}
		return &queryString, nil
	}
	return nil, errors.New("bad query syntax")

}
