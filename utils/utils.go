package utils

import (
	"errors"
	"net"
	"net/url"
	"regexp"
	serviceDiscovery "serviceDiscovery/consul"
	"strings"
)

const (
	consul string = "consul:"
)

// Endpoint represents the structure of an endpoint.
type DefaultHost struct {
	Host string
	Port string
}

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
	//consul:[dc1,dc2].namespace.serviceA.[tag1,tag2];abc.com:80
	queryAndDefaultHost := strings.Split(value, ";")
	if len(queryAndDefaultHost) != 2 {
		return nil, errors.New("default host not provided")
	}
	query := queryAndDefaultHost[0]
	//defaultHostString := queryAndDefaultHost[1]
	split := strings.Split(query, ":")
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

func cleanString(str string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	return reg.ReplaceAllString(str, "")
}

func getDefaultHost(str string) DefaultHost {
	val, err1 := url.Parse(str)

	defHost := DefaultHost{
		Host: "",
		Port: "",
	}
	if err1 != nil {
		//try with SplitHostPort
		h, p, err2 := net.SplitHostPort(str)
		if err2 != nil {
			//try with ParseIP
			ip := net.ParseIP(str)
			if ip != nil {
				defHost.Host = ip.String()
			}
		} else {
			defHost.Host = h
			defHost.Port = p
		}
	} else {
		if val != nil {
			defHost.Host = val.Hostname()
			defHost.Port = val.Port()
		}
	}

	return defHost
}
