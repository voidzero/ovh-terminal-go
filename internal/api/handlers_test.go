// internal/api/handlers_test.go
package api

import (
	"encoding/json"
	"testing"

	"ovh-terminal/internal/config"
	"ovh-terminal/internal/logger"
)

// mockClient simulates API responses for testing
type mockClient struct {
	responses map[string]interface{}
	errors    map[string]error
}

func (m *mockClient) Get(path string, result interface{}) error {
	if err, exists := m.errors[path]; exists && err != nil {
		return err
	}

	if response, exists := m.responses[path]; exists {
		// Simulate JSON marshaling/unmarshaling to match real behavior
		data, err := json.Marshal(response)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, result)
	}

	return nil
}

func (m *mockClient) Post(path string, payload interface{}, result interface{}) error {
	// Similar to Get, but for POST requests (to be implemented when needed)
	return nil
}

// Test data
var mockAccountInfo = &AccountInfo{
	Email:        "test@example.com",
	FirstName:    "Test",
	Name:         "User",
	Currency:     &Currency{Code: "EUR", Symbol: "€"},
	CustomerCode: "cc123456",
	NicHandle:    "ab12345-ovh",
}

var mockServerInfo = &ServerInfo{
	Name:         "ns123456.ip-1-2-3.eu",
	DisplayName:  "My Server",
	IP:           "1.2.3.4",
	State:        "active",
	Datacenter:   "rbx1",
	SupportLevel: "premium",
}

var mockDomainInfo = &DomainInfo{
	Domain:       "example.com",
	NameServers:  []string{"ns1.ovh.net", "ns2.ovh.net"},
	DnssecStatus: "active",
	WhoisOwner:   "Test User",
}

func setupMockClient() *Client {
	mock := &mockClient{
		responses: map[string]interface{}{
			"/me":                       mockAccountInfo,
			"/dedicated/server":         []string{"server1", "server2"},
			"/dedicated/server/server1": mockServerInfo,
			"/domain":                   []string{"example.com", "example.org"},
			"/domain/example.com":       mockDomainInfo,
			"/cloud/project":            []string{"project1", "project2"},
			"/ip":                       []string{"1.2.3.4", "5.6.7.8"},
			"/ip/1.2.3.4":               &IPInfo{IP: "1.2.3.4", Type: "failover"},
		},
		errors: make(map[string]error),
	}

	return &Client{
		client: mock,
		logger: logger.NewLogger(),
	}
}

func TestGetAccountInfo(t *testing.T) {
	client := setupMockClient()

	info, err := client.GetAccountInfo()
	if err != nil {
		t.Errorf("GetAccountInfo failed: %v", err)
	}

	if info.Email != mockAccountInfo.Email {
		t.Errorf("Expected email %s, got %s", mockAccountInfo.Email, info.Email)
	}
	if info.CustomerCode != mockAccountInfo.CustomerCode {
		t.Errorf(
			"Expected customer code %s, got %s",
			mockAccountInfo.CustomerCode,
			info.CustomerCode,
		)
	}
}

func TestListDedicatedServers(t *testing.T) {
	client := setupMockClient()

	servers, err := client.ListDedicatedServers()
	if err != nil {
		t.Errorf("ListDedicatedServers failed: %v", err)
	}

	expectedCount := 2
	if len(servers) != expectedCount {
		t.Errorf("Expected %d servers, got %d", expectedCount, len(servers))
	}
}

func TestGetDedicatedServerInfo(t *testing.T) {
	client := setupMockClient()

	info, err := client.GetDedicatedServerInfo("server1")
	if err != nil {
		t.Errorf("GetDedicatedServerInfo failed: %v", err)
	}

	if info.Name != mockServerInfo.Name {
		t.Errorf("Expected server name %s, got %s", mockServerInfo.Name, info.Name)
	}
	if info.State != mockServerInfo.State {
		t.Errorf("Expected server state %s, got %s", mockServerInfo.State, info.State)
	}
}

func TestListDomains(t *testing.T) {
	client := setupMockClient()

	domains, err := client.ListDomains()
	if err != nil {
		t.Errorf("ListDomains failed: %v", err)
	}

	expectedCount := 2
	if len(domains) != expectedCount {
		t.Errorf("Expected %d domains, got %d", expectedCount, len(domains))
	}
}

func TestGetDomainInfo(t *testing.T) {
	client := setupMockClient()

	info, err := client.GetDomainInfo("example.com")
	if err != nil {
		t.Errorf("GetDomainInfo failed: %v", err)
	}

	if info.Domain != mockDomainInfo.Domain {
		t.Errorf("Expected domain %s, got %s", mockDomainInfo.Domain, info.Domain)
	}
	if len(info.NameServers) != len(mockDomainInfo.NameServers) {
		t.Errorf("Expected %d nameservers, got %d",
			len(mockDomainInfo.NameServers), len(info.NameServers))
	}
}

func TestErrorHandling(t *testing.T) {
	client := setupMockClient()

	// Add an error for a specific path
	mock := client.client.(*mockClient)
	mock.errors["/me"] = NewAPIError("Test error", nil, nil)

	_, err := client.GetAccountInfo()
	if err == nil {
		t.Error("Expected error but got nil")
	}

	// Test error type
	if _, ok := err.(*APIError); !ok {
		t.Errorf("Expected APIError but got %T", err)
	}
}
