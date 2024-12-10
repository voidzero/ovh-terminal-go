// internal/api/endpoints.go
package api

import "fmt"

// Endpoint definitions for OVH API
const (
    // Account endpoints
    EndpointMe = "/me"

    // Server endpoints
    EndpointDedicatedServers = "/dedicated/server"

    // Domain endpoints
    EndpointDomains = "/domain"

    // Cloud endpoints
    EndpointCloudProjects = "/cloud/project"

    // IP endpoints
    EndpointIPs = "/ip"
)

// EndpointBuilder helps construct endpoint paths
type EndpointBuilder struct {
    base string
}

func NewEndpointBuilder(base string) *EndpointBuilder {
    return &EndpointBuilder{base: base}
}

func (eb *EndpointBuilder) WithID(id string) *EndpointBuilder {
    if id != "" {
        eb.base = fmt.Sprintf("%s/%s", eb.base, id)
    }
    return eb
}

func (eb *EndpointBuilder) WithAction(action string) *EndpointBuilder {
    if action != "" {
        eb.base = fmt.Sprintf("%s/%s", eb.base, action)
    }
    return eb
}

func (eb *EndpointBuilder) String() string {
    return eb.base
}

// Helper functions for common endpoint patterns
func DedicatedServerEndpoint(serverID string) string {
    return NewEndpointBuilder(EndpointDedicatedServers).WithID(serverID).String()
}

func DomainEndpoint(domain string) string {
    return NewEndpointBuilder(EndpointDomains).WithID(domain).String()
}

func CloudProjectEndpoint(projectID string) string {
    return NewEndpointBuilder(EndpointCloudProjects).WithID(projectID).String()
}