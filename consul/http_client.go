package serviceDiscovery

import (
	"context"
	"fmt"
	"github.com/hashicorp/consul/api"
)

//todo import loggers



// Data for a service instance
type NodeInfo struct {
	ServiceName string
	DataCenter  string
	Address     string
	ServicePort int
	ServiceTags []string
	ModifyIndex uint64   //whether the result has altered from the previous query
}

type Query struct {
	Datacenters  []string
	ServiceName  string
	Namespace    string
	Tags         []string
	QueryOptions *api.QueryOptions
}

// wraps the official go consul client
type consulClient struct {
	api *api.Client
}

type ConsulClient interface {
	// all services
	Services(q *api.QueryOptions) (map[string][]string, *api.QueryMeta, error)
	//get all nodes
	Nodes(ctx context.Context, query *Query, resultChan chan *NodeInfo, errorChan chan error)
	// single service
	Service(service, tag string, q *api.QueryOptions) ([]*api.CatalogService, *api.QueryMeta, error)
	// mesh scenario,
	Connect(service, tag string, q *api.QueryOptions) ([]*api.CatalogService, *api.QueryMeta, error)
}

//constructor
func NewConsulClient(a *api.Client) (*consulClient, error) {
	return &consulClient{api: a}, nil
}

// wraps the methods in the official api
func (c *consulClient) Nodes(ctx context.Context, query *Query, resultChan chan *NodeInfo, errorChan chan error)  {
	//todo optimize: reduce the no of calls
	//resultChan := make(chan NodeInfo)
	//errorChan := make(chan error)
	for _, dc := range query.Datacenters {
		qo := query.QueryOptions.WithContext(ctx) // returns a new obj with ctx
		qo.Datacenter = dc
		qo.Namespace = query.Namespace
		for _, tag := range query.Tags {
			go func(datacenter string, serviceName string, tag string, qo *api.QueryOptions) {
				fmt.Println("Go routine started", serviceName, tag, datacenter)
				defer fmt.Println("Go routine exited", serviceName, tag, datacenter)
				res, _, err := c.api.Catalog().Service(query.ServiceName, tag, qo)
				if err != nil {
					errorChan <- err
				}
				for _, r := range res {
					nodeInfo := NodeInfo{
						ServiceName: r.ServiceName,
						DataCenter:  r.Datacenter,
						Address:     r.Address,
						ServicePort: r.ServicePort,
						ServiceTags: r.ServiceTags,
						ModifyIndex: r.ModifyIndex,
					}
					resultChan <- &nodeInfo
				}
			}(dc, query.ServiceName, tag, qo)
		}
	}
}

func (c *consulClient) Services(q *api.QueryOptions) (map[string][]string, *api.QueryMeta, error) {
	return c.api.Catalog().Services(q)
}

func (c *consulClient) Service(service, tag string, q *api.QueryOptions) ([]*api.CatalogService, *api.QueryMeta, error) {
	p, s, r := c.api.Catalog().Service(service, tag, q)
	//for _, i := range p {
	//	fmt.Println(i)
	//}
	return p, s, r
}

func (c *consulClient) Connect(service, tag string, q *api.QueryOptions) ([]*api.CatalogService, *api.QueryMeta, error) {
	return c.api.Catalog().Connect(service, tag, q)
}
