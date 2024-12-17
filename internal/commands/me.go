// internal/commands/me.go
package commands

import (
	"context"
	"fmt"
	"time"

	"ovh-terminal/internal/api"
	"ovh-terminal/internal/format"
	"ovh-terminal/internal/logger"
)

// Configuration constants
const (
	keyValueSpacing = 4  // Extra spacing between key and value
	maxWidth        = 80 // Maximum width of the output
)

// SectionOrder defines the display order
var SectionOrder = []string{"account", "company", "personal", "address"}

// sectionTitles maps section identifiers to display titles
var sectionTitles = map[string]string{
	"account":  "Account Details",
	"company":  "Company Information",
	"personal": "Personal Information",
	"address":  "Address",
}

// MeCommand handles the account info display
type MeCommand struct {
	BaseCommand
	client *api.Client
	log    *logger.Logger
}

// NewMeCommand creates a new me command instance
func NewMeCommand(client *api.Client) *MeCommand {
	return &MeCommand{
		BaseCommand: NewBaseCommand(TypeInfo),
		client:      client,
		log:         logger.Log.With(map[string]interface{}{"command": "me"}),
	}
}

// Execute implements the Command interface
func (c *MeCommand) Execute() (string, error) {
	return c.ExecuteWithOptions()
}

// ExecuteWithOptions implements the Command interface
func (c *MeCommand) ExecuteWithOptions(opts ...CommandOption) (string, error) {
	for _, opt := range opts {
		opt(&c.config)
	}
	return c.executeWithTimeout(context.Background(), c.executeCommand)
}

// ExecuteAsync implements the Command interface
func (c *MeCommand) ExecuteAsync(ctx context.Context) (<-chan CommandResult, error) {
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

// executeCommand handles the actual command execution
func (c *MeCommand) executeCommand() (string, error) {
	c.log.Debug("Executing me command")

	info, err := c.client.GetAccountInfo()
	if err != nil {
		c.log.Error("Failed to get account info", "error", err)
		return "", fmt.Errorf("failed to get account info: %w", err)
	}

	output := format.NewOutputFormatter(
		format.WithMaxWidth(maxWidth),
		format.WithSeparator("\n"),
	)

	// Create sections in defined order
	for _, name := range SectionOrder {
		section := output.AddSection(sectionTitles[name])
		section.SetConfig(format.SectionConfig{
			KeyValueSpacing: keyValueSpacing,
			TitleDecorator:  "=",
		})

		switch name {
		case "account":
			formatAccountSection(info, section)
		case "company":
			formatCompanySection(info, section)
		case "personal":
			formatPersonalSection(info, section)
		case "address":
			formatAddressSection(info, section)
		}
	}

	return output.String(), nil
}

// Section formatters
func formatAccountSection(info *api.AccountInfo, section *format.Section) {
	section.AddField("NIC Handle", info.NicHandle)
	section.AddField("Customer Code", info.CustomerCode)
	section.AddField("Account State", info.State)
	section.AddField("KYC Validated", fmt.Sprintf("%v", info.KYCValidated))
}

func formatCompanySection(info *api.AccountInfo, section *format.Section) {
	if info.Organisation != "" {
		section.AddField("Organization", info.Organisation)
	}
	if info.Currency != nil {
		section.AddField("Currency", fmt.Sprintf("%s (%s)",
			info.Currency.Code, info.Currency.Symbol))
	}
}

func formatPersonalSection(info *api.AccountInfo, section *format.Section) {
	section.AddField("Name", fmt.Sprintf("%s %s", info.FirstName, info.Name))
	section.AddField("Email", info.Email)
	if info.Phone != "" {
		phone := info.Phone
		if info.PhoneCountry != "" {
			phone = fmt.Sprintf("%s (%s)", phone, info.PhoneCountry)
		}
		section.AddField("Phone", phone)
	}
	section.AddField("Language", info.Language)
}

func formatAddressSection(info *api.AccountInfo, section *format.Section) {
	section.AddField("Street", info.Address)
	section.AddField("Postal Code", info.ZIP)
	section.AddField("City", info.City)
	section.AddField("Country", info.Country)
}
