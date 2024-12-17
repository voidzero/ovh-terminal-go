// internal/api/endpoints.go
package api

import (
	"fmt"
	"path"
	"strings"
)

// ResourceType represents different API resource types
type ResourceType string

const (
	ResourceAccount ResourceType = "account"
	ResourceServer  ResourceType = "server"
	ResourceDomain  ResourceType = "domain"
	ResourceCloud   ResourceType = "cloud"
	ResourceIP      ResourceType = "ip"
	ResourceBilling ResourceType = "billing"
	ResourceSupport ResourceType = "support"
)

// Endpoint definitions for OVH API
const (
	// Base endpoints
	endpointMe              = "/me"
	endpointDedicatedServer = "/dedicated/server"
	endpointDomain          = "/domain"
	endpointCloudProject    = "/cloud/project"
	endpointIP              = "/ip"
	endpointBilling         = "/me/bill"
	endpointSupport         = "/support"
)

// EndpointMap maps resource types to their base endpoints
var EndpointMap = map[ResourceType]string{
	ResourceAccount: endpointMe,
	ResourceServer:  endpointDedicatedServer,
	ResourceDomain:  endpointDomain,
	ResourceCloud:   endpointCloudProject,
	ResourceIP:      endpointIP,
	ResourceBilling: endpointBilling,
	ResourceSupport: endpointSupport,
}

// EndpointBuilder helps construct endpoint paths
type EndpointBuilder struct {
	base       string
	segments   []string
	parameters map[string]string
}

// NewEndpointBuilder creates a new endpoint builder
func NewEndpointBuilder(resourceType ResourceType) *EndpointBuilder {
	base, ok := EndpointMap[resourceType]
	if !ok {
		base = "/" + string(resourceType)
	}

	return &EndpointBuilder{
		base:       base,
		segments:   make([]string, 0),
		parameters: make(map[string]string),
	}
}

// WithID adds a resource ID to the path
func (eb *EndpointBuilder) WithID(id string) *EndpointBuilder {
	if id != "" {
		eb.segments = append(eb.segments, id)
	}
	return eb
}

// WithSegment adds a path segment
func (eb *EndpointBuilder) WithSegment(segment string) *EndpointBuilder {
	if segment != "" {
		eb.segments = append(eb.segments, segment)
	}
	return eb
}

// WithAction adds an action to the path
func (eb *EndpointBuilder) WithAction(action string) *EndpointBuilder {
	return eb.WithSegment(action)
}

// WithParameter adds a query parameter
func (eb *EndpointBuilder) WithParameter(key, value string) *EndpointBuilder {
	if key != "" && value != "" {
		eb.parameters[key] = value
	}
	return eb
}

// Build constructs the final endpoint URL
func (eb *EndpointBuilder) Build() string {
	// Start with base path
	fullPath := eb.base

	// Add segments
	if len(eb.segments) > 0 {
		segments := make([]string, len(eb.segments))
		for i, seg := range eb.segments {
			segments[i] = strings.TrimLeft(seg, "/")
		}
		fullPath = path.Join(fullPath, path.Join(segments...))
	}

	// Add query parameters
	if len(eb.parameters) > 0 {
		params := make([]string, 0, len(eb.parameters))
		for k, v := range eb.parameters {
			params = append(params, fmt.Sprintf("%s=%s", k, v))
		}
		fullPath = fullPath + "?" + strings.Join(params, "&")
	}

	return fullPath
}

// String implements Stringer interface
func (eb *EndpointBuilder) String() string {
	return eb.Build()
}

// Helper functions for common endpoint patterns
func GetAccountEndpoint() string {
	return NewEndpointBuilder(ResourceAccount).Build()
}

func GetServerEndpoint(serverID string) string {
	return NewEndpointBuilder(ResourceServer).WithID(serverID).Build()
}

func GetServerActionEndpoint(serverID, action string) string {
	return NewEndpointBuilder(ResourceServer).
		WithID(serverID).
		WithAction(action).
		Build()
}

func GetDomainEndpoint(domain string) string {
	return NewEndpointBuilder(ResourceDomain).WithID(domain).Build()
}

func GetDomainActionEndpoint(domain, action string) string {
	return NewEndpointBuilder(ResourceDomain).
		WithID(domain).
		WithAction(action).
		Build()
}

func GetCloudProjectEndpoint(projectID string) string {
	return NewEndpointBuilder(ResourceCloud).WithID(projectID).Build()
}

func GetCloudProjectActionEndpoint(projectID, action string) string {
	return NewEndpointBuilder(ResourceCloud).
		WithID(projectID).
		WithAction(action).
		Build()
}

func GetIPEndpoint(ip string) string {
	return NewEndpointBuilder(ResourceIP).WithID(ip).Build()
}

func GetBillingEndpoint(billID string) string {
	return NewEndpointBuilder(ResourceBilling).WithID(billID).Build()
}

func GetSupportTicketEndpoint(ticketID string) string {
	return NewEndpointBuilder(ResourceSupport).WithID(ticketID).Build()
}

