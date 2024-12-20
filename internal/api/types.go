// internal/api/types.go
package api

import (
	"fmt"
	"strings"
	"time"
)

// AccountInfo represents the /me endpoint response
type AccountInfo struct {
	Email        string    `json:"email"`
	FirstName    string    `json:"firstname"`
	Name         string    `json:"name"`
	Currency     *Currency `json:"currency"`
	Phone        string    `json:"phone"`
	PhoneCountry string    `json:"phoneCountry,omitempty"`
	SpareEmail   string    `json:"spareEmail,omitempty"`
	Language     string    `json:"language"`
	Organisation string    `json:"organisation,omitempty"`
	City         string    `json:"city,omitempty"`
	Address      string    `json:"address,omitempty"`
	ZIP          string    `json:"zip,omitempty"`
	Country      string    `json:"country,omitempty"`
	CustomerCode string    `json:"customerCode"`
	NicHandle    string    `json:"nichandle"`
	State        string    `json:"state"`
	KYCValidated bool      `json:"kycValidated"`
}

// GetFullName returns the combined first and last name
func (a *AccountInfo) GetFullName() string {
	return strings.TrimSpace(fmt.Sprintf("%s %s", a.FirstName, a.Name))
}

// GetFormattedPhone returns a formatted phone number with country code
func (a *AccountInfo) GetFormattedPhone() string {
	if a.Phone == "" {
		return ""
	}
	if a.PhoneCountry != "" {
		return fmt.Sprintf("%s (%s)", a.Phone, a.PhoneCountry)
	}
	return a.Phone
}

// GetFormattedAddress returns a formatted complete address
func (a *AccountInfo) GetFormattedAddress() string {
	parts := []string{}
	if a.Address != "" {
		parts = append(parts, a.Address)
	}
	if a.ZIP != "" || a.City != "" {
		parts = append(parts, strings.TrimSpace(fmt.Sprintf("%s %s", a.ZIP, a.City)))
	}
	if a.Country != "" {
		parts = append(parts, a.Country)
	}
	return strings.Join(parts, "\n")
}

// Currency represents a currency with code and symbol
type Currency struct {
	Code   string `json:"code"`
	Symbol string `json:"symbol"`
}

// String implements Stringer interface for Currency
func (c *Currency) String() string {
	if c == nil {
		return ""
	}
	return fmt.Sprintf("%s (%s)", c.Code, c.Symbol)
}

// ServerState represents possible server states
type ServerState string

const (
	ServerStateActive      ServerState = "active"
	ServerStateInactive    ServerState = "inactive"
	ServerStateMaintenance ServerState = "maintenance"
)

// IAMInfo represents IAM-related server information
type IAMInfo struct {
	DisplayName string `json:"displayName"`
	ID          string `json:"id"`
	URN         string `json:"urn"`
}

// ServerInfo represents dedicated server information
type ServerInfo struct {
	Name             string      `json:"name"`
	IP               string      `json:"ip"`
	State            ServerState `json:"state"`
	PowerState       string      `json:"powerState"`
	Datacenter       string      `json:"datacenter"`
	SupportLevel     string      `json:"supportLevel"`
	Professional     bool        `json:"professionalUse"`
	CommercialRange  string      `json:"commercialRange"`
	Reverse          string      `json:"reverse"`
	AvailabilityZone string      `json:"availabilityZone"`
	Region           string      `json:"region"`
	Rack             string      `json:"rack"`
	OS               string      `json:"os"`
	Monitoring       bool        `json:"monitoring"`
	LinkSpeed        int         `json:"linkSpeed"`
	NoIntervention   bool        `json:"noIntervention"`
	IAM              *IAMInfo    `json:"iam"`
} // IsOperational checks if the server is in a working state
func (s *ServerInfo) IsOperational() bool {
	return s.State == ServerStateActive
}

// GetDisplayTitle returns the best available name for the server
func (s *ServerInfo) GetDisplayTitle() string {
	// First try IAM display name
	if s.IAM != nil && s.IAM.DisplayName != "" {
		return s.IAM.DisplayName
	}

	// Then try reverse DNS (without trailing dot)
	if s.Reverse != "" {
		if s.Reverse[len(s.Reverse)-1] == '.' {
			return s.Reverse[:len(s.Reverse)-1]
		}
		return s.Reverse
	}

	// Finally fall back to server name
	return s.Name
}

// DomainInfo represents domain information
type DomainInfo struct {
	Domain       string    `json:"domain"`
	NameServers  []string  `json:"nameServers"`
	DnssecStatus string    `json:"dnssecStatus"`
	LastUpdate   string    `json:"lastUpdate"`
	WhoisOwner   string    `json:"whoisOwner"`
	Expiration   time.Time `json:"expiration"`
}

// IsExpired checks if the domain has expired
func (d *DomainInfo) IsExpired() bool {
	return time.Now().After(d.Expiration)
}

// ExpiresWithin checks if the domain expires within the given duration
func (d *DomainInfo) ExpiresWithin(duration time.Duration) bool {
	return time.Now().Add(duration).After(d.Expiration)
}

// GetFormattedNameServers returns a formatted list of nameservers
func (d *DomainInfo) GetFormattedNameServers() string {
	return strings.Join(d.NameServers, ", ")
}

// IPType represents different types of IP addresses
type IPType string

const (
	IPTypeFailover IPType = "failover"
	IPTypeCloud    IPType = "cloud"
	IPTypeVPS      IPType = "vps"
)

// IPInfo represents IP information
type IPInfo struct {
	IP          string   `json:"ip"`
	Type        IPType   `json:"type"`
	Description string   `json:"description"`
	RoutedTo    string   `json:"routedTo"`
	IPBlocks    []string `json:"ipBlock"`
}

// IsFailover checks if this is a failover IP
func (i *IPInfo) IsFailover() bool {
	return i.Type == IPTypeFailover
}

// GetFormattedType returns a human-readable IP type
func (i *IPInfo) GetFormattedType() string {
	switch i.Type {
	case IPTypeFailover:
		return "Failover IP"
	case IPTypeCloud:
		return "Cloud IP"
	case IPTypeVPS:
		return "VPS IP"
	default:
		return string(i.Type)
	}
}

// GetFormattedDescription returns a description or default text
func (i *IPInfo) GetFormattedDescription() string {
	if i.Description != "" {
		return i.Description
	}
	return "No description available"
}

// VPSModelInfo represents VPS model information
type VPSModelInfo struct {
	MaximumAdditionnalIp int      `json:"maximumAdditionnalIp"`
	AvailableOptions     []string `json:"availableOptions"`
	Offer                string   `json:"offer"`
	Name                 string   `json:"name"`
	Datacenter           []string `json:"datacenter"`
	Memory               int      `json:"memory"`
	Version              string   `json:"version"`
	Disk                 int      `json:"disk"`
	VCore                int      `json:"vcore"`
}

// VPSInfo represents VPS information from the API
type VPSInfo struct {
	SLAMonitoring      bool         `json:"slaMonitoring"`
	Zone               string       `json:"zone"`
	State              string       `json:"state"`
	OfferType          string       `json:"offerType"`
	MemoryLimit        int          `json:"memoryLimit"`
	MonitoringIPBlocks []string     `json:"monitoringIpBlocks"`
	Model              VPSModelInfo `json:"model"`
	VCore              int          `json:"vcore"`
	Cluster            string       `json:"cluster"`
	DisplayName        string       `json:"displayName"`
	Keymap             *string      `json:"keymap"`
	NetbootMode        string       `json:"netbootMode"`
	Name               string       `json:"name"`
	IAM                *IAMInfo     `json:"iam"`
}

// GetDisplayTitle returns the best available name for the VPS
func (v *VPSInfo) GetDisplayTitle() string {
	if v.DisplayName != "" {
		return v.DisplayName
	}
	return v.Name
}

// IsOperational checks if the VPS is in a working state
func (v *VPSInfo) IsOperational() bool {
	return v.State == "running"
}
