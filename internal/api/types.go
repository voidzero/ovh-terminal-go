// internal/api/types.go
package api

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

type Currency struct {
	Code   string `json:"code"`
	Symbol string `json:"symbol"`
}

// ServerInfo represents dedicated server information
type ServerInfo struct {
	Name         string `json:"name"`
	DisplayName  string `json:"displayName"`
	IP           string `json:"ip"`
	State        string `json:"state"`
	Datacenter   string `json:"datacenter"`
	SupportLevel string `json:"supportLevel"`
	Professional bool   `json:"professional"`
}

// DomainInfo represents domain information
type DomainInfo struct {
	Domain       string   `json:"domain"`
	NameServers  []string `json:"nameServers"`
	DnssecStatus string   `json:"dnssecStatus"`
	LastUpdate   string   `json:"lastUpdate"`
	WhoisOwner   string   `json:"whoisOwner"`
}

// IPInfo represents IP information
type IPInfo struct {
	IP          string   `json:"ip"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	RoutedTo    string   `json:"routedTo"`
	IpBlocks    []string `json:"ipBlock"`
}
