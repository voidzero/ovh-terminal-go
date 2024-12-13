package commands

import (
	"fmt"

	"ovh-terminal/internal/api"
	"ovh-terminal/internal/format"
)

// APIInfoCommand handles the API applications and credentials info display
type APIInfoCommand struct {
	client *api.Client
}

// NewAPIInfoCommand creates a new API info command instance
func NewAPIInfoCommand(client *api.Client) *APIInfoCommand {
	return &APIInfoCommand{
		client: client,
	}
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

// Execute runs the command
func (c *APIInfoCommand) Execute() (string, error) {
	var appIDs []int
	var credIDs []int

	// Get list of application IDs
	if err := c.client.Get("/me/api/application", &appIDs); err != nil {
		return "", fmt.Errorf("failed to fetch application IDs: %w", err)
	}

	// Get list of credential IDs
	if err := c.client.Get("/me/api/credential", &credIDs); err != nil {
		return "", fmt.Errorf("failed to fetch credential IDs: %w", err)
	}

	// Fetch all credentials details and build a map of all known application IDs
	creds := make(map[int]Credential)
	knownAppIDs := make(map[int]bool)
	for _, credID := range credIDs {
		var cred Credential
		if err := c.client.Get(fmt.Sprintf("/me/api/credential/%d", credID), &cred); err != nil {
			return "", fmt.Errorf("failed to fetch credential %d details: %w", credID, err)
		}
		creds[credID] = cred
		knownAppIDs[cred.ApplicationID] = true
	}

	// Create maps for easier lookup
	credsByApp := make(map[int][]Credential)
	for _, cred := range creds {
		credsByApp[cred.ApplicationID] = append(credsByApp[cred.ApplicationID], cred)
	}

	// Format the output using the formatter
	output := format.NewOutputFormatter()

	// First handle known applications from the API
	for _, appID := range appIDs {
		var app Application
		if err := c.client.Get(fmt.Sprintf("/me/api/application/%d", appID), &app); err != nil {
			return "", fmt.Errorf("failed to fetch application %d details: %w", appID, err)
		}

		appSection := output.AddSection(fmt.Sprintf("Application: %s", app.Name))

		appSection.AddField("ID", fmt.Sprintf("%d", app.ApplicationID))
		appSection.AddField("API Key", app.ApplicationKey)
		appSection.AddField("Description", app.Description)
		appSection.AddField("Status", app.Status)

		if appCreds, exists := credsByApp[app.ApplicationID]; exists {
			for _, cred := range appCreds {
				appSection.AddField(
					fmt.Sprintf("\nCredential %d", cred.CredentialID),
					formatCredential(cred),
				)
			}
			delete(credsByApp, app.ApplicationID)
			delete(knownAppIDs, app.ApplicationID)
		} else {
			appSection.AddField("Credentials", "No active credentials")
		}
	}

	// Handle special cases and unknown applications
	for appID := range knownAppIDs {
		var appName, appDesc string
		switch appID {
		case 115:
			appName = "OVH Website"
			appDesc = "Official OVH website application"
		default:
			appName = "Unknown Application"
			appDesc = "Application details not available"
		}

		appSection := output.AddSection(fmt.Sprintf("Application: %s", appName))
		appSection.AddField("ID", fmt.Sprintf("%d", appID))
		appSection.AddField("Description", appDesc)

		if appCreds, exists := credsByApp[appID]; exists {
			for _, cred := range appCreds {
				appSection.AddField(
					fmt.Sprintf("Credential %d", cred.CredentialID),
					formatCredential(cred),
				)
			}
			delete(credsByApp, appID)
		}
	}

	return output.String(), nil
}

// Helper function to format a credential
func formatCredential(cred Credential) string {
	var details string
	details = fmt.Sprintf("\nStatus: %s\nCreated: %s\nExpires: %s\nLast used: %s",
		cred.Status,
		cred.Creation,
		cred.Expiration,
		cred.LastUse,
	)

	if len(cred.AllowedIPs) > 0 {
		details += "\nAllowed IPs:"
		for _, ip := range cred.AllowedIPs {
			details += fmt.Sprintf("\n• %s", ip)
		}
	}

	if len(cred.Rules) > 0 {
		details += "\nAccess rules:"
		for _, rule := range cred.Rules {
			details += fmt.Sprintf("\n• %s %s", rule.Method, rule.Path)
		}
	}

	if cred.OVHSupport {
		details += "\nOVH Support access enabled"
	}

	return details
}

