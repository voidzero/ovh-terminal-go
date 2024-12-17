// internal/commands/api_info.go
package commands

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"ovh-terminal/internal/api"
	"ovh-terminal/internal/format"
	"ovh-terminal/internal/logger"
)

// APIInfoCommand handles the API applications and credentials info display
type APIInfoCommand struct {
	BaseCommand
	client *api.Client
	log    *logger.Logger
}

// NewAPIInfoCommand creates a new API info command instance
func NewAPIInfoCommand(client *api.Client) *APIInfoCommand {
	return &APIInfoCommand{
		BaseCommand: NewBaseCommand(TypeInfo),
		client:      client,
		log:         logger.Log.With(map[string]interface{}{"command": "api_info"}),
	}
}

// Execute implements the Command interface
func (c *APIInfoCommand) Execute() (string, error) {
	return c.ExecuteWithOptions()
}

// ExecuteWithOptions implements the Command interface
func (c *APIInfoCommand) ExecuteWithOptions(opts ...CommandOption) (string, error) {
	// Apply options to base command
	for _, opt := range opts {
		opt(&c.config)
	}

	return c.executeWithTimeout(context.Background(), func() (string, error) {
		return c.executeCommand()
	})
}

// ExecuteAsync implements the Command interface
func (c *APIInfoCommand) ExecuteAsync(ctx context.Context) (<-chan CommandResult, error) {
	resultCh := make(chan CommandResult, 1)

	go func() {
		defer close(resultCh)

		start := time.Now()
		output, err := c.executeCommand()
		duration := time.Since(start)

		state := StateCompleted
		if err != nil {
			state = StateFailed
		}

		resultCh <- CommandResult{
			Output:   output,
			Error:    err,
			Duration: duration,
			State:    state,
		}
	}()

	return resultCh, nil
}

// Application represents an OVH API application
type Application struct {
	ApplicationID  int    `json:"applicationId"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Status         string `json:"status"`
	ApplicationKey string `json:"applicationKey"`
}

// Credential represents an OVH API credential
type Credential struct {
	CredentialID  int      `json:"credentialId"`
	ApplicationID int      `json:"applicationId"`
	Status        string   `json:"status"`
	LastUse       string   `json:"lastUse"`
	Creation      string   `json:"creation"`
	Expiration    string   `json:"expiration"`
	AllowedIPs    []string `json:"allowedIPs"`
	OVHSupport    bool     `json:"ovhSupport"`
	Rules         []Rule   `json:"rules"`
}

// Rule represents an API access rule
type Rule struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

// AppData represents organized application data
type AppData struct {
	App         Application
	Credentials []Credential
}

// executeCommand handles the actual command execution
func (c *APIInfoCommand) executeCommand() (string, error) {
	c.log.Debug("Executing api_info command")

	// Fetch applications and credentials
	apps, creds, err := c.fetchData()
	if err != nil {
		return "", err
	}

	// Create organized data structure
	data := c.organizeData(apps, creds)

	// Format output
	return c.formatOutput(data), nil
}

// fetchData retrieves all necessary data from the API
func (c *APIInfoCommand) fetchData() (map[int]Application, map[int]Credential, error) {
	var appIDs []int
	if err := c.client.Get("/me/api/application", &appIDs); err != nil {
		c.log.Error("Failed to fetch application IDs", "error", err)
		return nil, nil, err
	}

	var credIDs []int
	if err := c.client.Get("/me/api/credential", &credIDs); err != nil {
		c.log.Error("Failed to fetch credential IDs", "error", err)
		return nil, nil, err
	}

	apps := make(map[int]Application)
	for _, id := range appIDs {
		var app Application
		if err := c.client.Get(fmt.Sprintf("/me/api/application/%d", id), &app); err != nil {
			c.log.Error("Failed to fetch application details", "id", id, "error", err)
			continue
		}
		apps[id] = app
	}

	creds := make(map[int]Credential)
	for _, id := range credIDs {
		var cred Credential
		if err := c.client.Get(fmt.Sprintf("/me/api/credential/%d", id), &cred); err != nil {
			c.log.Error("Failed to fetch credential details", "id", id, "error", err)
			continue
		}
		creds[id] = cred
	}

	return apps, creds, nil
}

// organizeData organizes applications and credentials into a structured format
func (c *APIInfoCommand) organizeData(
	apps map[int]Application,
	creds map[int]Credential,
) map[int]AppData {
	data := make(map[int]AppData)

	// Initialize with known applications
	for id, app := range apps {
		data[id] = AppData{
			App:         app,
			Credentials: make([]Credential, 0),
		}
	}

	// Add special cases for known application IDs
	specialApps := map[int]Application{
		115: {
			ApplicationID: 115,
			Name:          "OVH Website",
			Description:   "Official OVH website application",
			Status:        "active",
		},
	}

	// Add credentials to their applications
	for _, cred := range creds {
		appID := cred.ApplicationID
		appData, exists := data[appID]
		if !exists {
			// Check if it's a special app
			if specialApp, ok := specialApps[appID]; ok {
				appData = AppData{
					App:         specialApp,
					Credentials: make([]Credential, 0),
				}
			} else {
				appData = AppData{
					App: Application{
						ApplicationID: appID,
						Name:          "Unknown Application",
						Description:   "Application details not available",
						Status:        "unknown",
					},
					Credentials: make([]Credential, 0),
				}
			}
			data[appID] = appData
		}
		appData.Credentials = append(appData.Credentials, cred)
		data[appID] = appData
	}

	return data
}

// formatOutput creates the formatted output string
func (c *APIInfoCommand) formatOutput(data map[int]AppData) string {
	output := format.NewOutputFormatter(
		format.WithMaxWidth(100),
		format.WithSeparator("\n"),
	)

	// Sort applications by ID
	var appIDs []int
	for id := range data {
		appIDs = append(appIDs, id)
	}
	sort.Ints(appIDs)

	// Format each application
	for _, appID := range appIDs {
		appData := data[appID]
		section := output.AddSection("Application: " + appData.App.Name)

		// Application details
		section.AddFields(map[string]string{
			"ID":          fmt.Sprintf("%d", appData.App.ApplicationID),
			"API Key":     appData.App.ApplicationKey,
			"Status":      appData.App.Status,
			"Description": appData.App.Description,
		})

		// Add credentials
		if len(appData.Credentials) == 0 {
			section.AddField("Credentials", "No active credentials")
		} else {
			for _, cred := range appData.Credentials {
				section.AddField(
					fmt.Sprintf("Credential %d", cred.CredentialID),
					formatCredential(cred),
				)
			}
		}
	}

	return output.String()
}

// Helper function to format a credential
func formatCredential(cred Credential) string {
	var details strings.Builder

	details.WriteString("\nStatus: " + cred.Status)
	details.WriteString("\nCreated: " + cred.Creation)
	if cred.Expiration != "" {
		details.WriteString("\nExpires: " + cred.Expiration)
	} else {
		details.WriteString("\nExpires:")
	}

	if cred.LastUse != "" {
		details.WriteString("\nLast used: " + cred.LastUse)
	} else {
		details.WriteString("\nLast used:")
	}

	if len(cred.AllowedIPs) > 0 {
		details.WriteString("\nAllowed IPs:")
		for _, ip := range cred.AllowedIPs {
			details.WriteString("\n• " + ip)
		}
	}

	if len(cred.Rules) > 0 {
		details.WriteString("\n\nAccess rules:")
		for _, rule := range cred.Rules {
			details.WriteString("\n• " + rule.Method + " " + rule.Path)
		}
		details.WriteString("\n")
	}

	if cred.OVHSupport {
		details.WriteString("\nOVH Support access enabled\n")
	}

	return details.String()
}
