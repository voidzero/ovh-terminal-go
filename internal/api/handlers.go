// internal/api/handlers.go
package api

import (
    "fmt"
)

// GetAccountInfo retrieves account information
func (c *Client) GetAccountInfo() (*AccountInfo, error) {
    var info AccountInfo
    err := c.Get(EndpointMe, &info)
    if err != nil {
        return nil, err  // Pass through the original error
    }
    return &info, nil
}

// ListDedicatedServers retrieves all dedicated servers
func (c *Client) ListDedicatedServers() ([]string, error) {
    var serverIDs []string
    err := c.Get(EndpointDedicatedServers, &serverIDs)
    if err != nil {
        return nil, fmt.Errorf("failed to list servers: %w", err)
    }
    return serverIDs, nil
}

// GetDedicatedServerInfo retrieves information about a specific server
func (c *Client) GetDedicatedServerInfo(serverID string) (*ServerInfo, error) {
    var info ServerInfo
    endpoint := DedicatedServerEndpoint(serverID)
    err := c.Get(endpoint, &info)
    if err != nil {
        return nil, fmt.Errorf("failed to get server info for %s: %w", serverID, err)
    }
    return &info, nil
}

// ListDomains retrieves all domains
func (c *Client) ListDomains() ([]string, error) {
    var domains []string
    err := c.Get(EndpointDomains, &domains)
    if err != nil {
        return nil, fmt.Errorf("failed to list domains: %w", err)
    }
    return domains, nil
}

// GetDomainInfo retrieves information about a specific domain
func (c *Client) GetDomainInfo(domain string) (*DomainInfo, error) {
    var info DomainInfo
    endpoint := DomainEndpoint(domain)
    err := c.Get(endpoint, &info)
    if err != nil {
        return nil, fmt.Errorf("failed to get domain info for %s: %w", domain, err)
    }
    return &info, nil
}

// ListCloudProjects retrieves all cloud projects
func (c *Client) ListCloudProjects() ([]string, error) {
    var projects []string
    err := c.Get(EndpointCloudProjects, &projects)
    if err != nil {
        return nil, fmt.Errorf("failed to list cloud projects: %w", err)
    }
    return projects, nil
}

// ListIPs retrieves all IPs
func (c *Client) ListIPs() ([]string, error) {
    var ips []string
    err := c.Get(EndpointIPs, &ips)
    if err != nil {
        return nil, fmt.Errorf("failed to list IPs: %w", err)
    }
    return ips, nil
}

// GetIPInfo retrieves information about a specific IP
func (c *Client) GetIPInfo(ip string) (*IPInfo, error) {
    var info IPInfo
    endpoint := fmt.Sprintf("%s/%s", EndpointIPs, ip)
    err := c.Get(endpoint, &info)
    if err != nil {
        return nil, fmt.Errorf("failed to get IP info for %s: %w", ip, err)
    }
    return &info, nil
}
