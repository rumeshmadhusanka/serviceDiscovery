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
	ModifyIndex uint64 //whether the result has altered from the previous query
}

type QueryString struct {
	Datacenters []string
	ServiceName string
	Namespace   string
	Tags        []string
}

type Query struct {
	QString      QueryString
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
	// todo consider mesh scenario,
	Connect(service, tag string, q *api.QueryOptions) ([]*api.CatalogService, *api.QueryMeta, error)
}

//constructor
func NewConsulClient(a *api.Client) (*consulClient, error) {
	return &consulClient{api: a}, nil
}

// wraps the methods in the official api
func (c *consulClient) Nodes(ctx context.Context, query *Query, resultChan chan *NodeInfo, errorChan chan error) {
	//todo optimize: reduce the no of calls
	//todo health checks passing
	//resultChan := make(chan NodeInfo)
	//errorChan := make(chan error)
	for _, dc := range query.QString.Datacenters {
		qo := query.QueryOptions.WithContext(ctx) // returns a new obj with ctx
		qo.Datacenter = dc
		qo.Namespace = query.QString.Namespace
		for _, tag := range query.QString.Tags {
			go func(datacenter string, serviceName string, tag string, qo *api.QueryOptions) {
				fmt.Println("Go routine started", serviceName, tag, datacenter)
				defer fmt.Println("Go routine exited", serviceName, tag, datacenter)
				res, _, err := c.api.Catalog().Service(query.QString.ServiceName, tag, qo)
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
					//check whether health checks are passing
					fmt.Println("checks: ", r.Checks)
					if len(r.Checks) == 0 {
						fmt.Println("check empty")
						resultChan <- &nodeInfo
					} else {
						for _, chk := range r.Checks {
							if chk.Status == "passing" {
								resultChan <- &nodeInfo
							}
						}
					}
				}
			}(dc, query.QString.ServiceName, tag, qo)
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
