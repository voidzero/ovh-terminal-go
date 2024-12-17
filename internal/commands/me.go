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

// SectionFormatter defines a function that formats a section of output
type SectionFormatter func(*api.AccountInfo, *format.Section)

// MeCommand handles the account info display
type MeCommand struct {
	BaseCommand
	client     *api.Client
	formatters map[string]SectionFormatter
	log        *logger.Logger
}

// NewMeCommand creates a new me command instance
func NewMeCommand(client *api.Client) *MeCommand {
	cmd := &MeCommand{
		BaseCommand: NewBaseCommand(TypeInfo),
		client:      client,
		formatters:  make(map[string]SectionFormatter),
		log:         logger.Log.With(map[string]interface{}{"command": "me"}),
	}

	cmd.registerFormatters()
	return cmd
}

// Execute implements the Command interface
func (c *MeCommand) Execute() (string, error) {
	return c.ExecuteWithOptions()
}

// ExecuteWithOptions implements the Command interface
func (c *MeCommand) ExecuteWithOptions(opts ...CommandOption) (string, error) {
	// Apply options to base command
	for _, opt := range opts {
		opt(&c.config)
	}

	return c.executeWithTimeout(context.Background(), func() (string, error) {
		return c.executeCommand()
	})
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

	// Get account info
	info, err := c.client.GetAccountInfo()
	if err != nil {
		c.log.Error("Failed to get account info", "error", err)
		return "", fmt.Errorf("failed to get account info: %w", err)
	}

	// Create output formatter with appropriate width
	output := format.NewOutputFormatter(
		format.WithMaxWidth(80),
		format.WithSeparator("\n"),
	)

	// Apply each section formatter
	for name, formatter := range c.formatters {
		section := output.AddSection(getSectionTitle(name))
		formatter(info, section)
	}

	c.log.Debug("Me command completed successfully")
	return output.String(), nil
}

// getSectionTitle returns the display title for a section
func getSectionTitle(section string) string {
	titles := map[string]string{
		"personal": "Personal Information",
		"company":  "Company Information",
		"address":  "Address",
		"account":  "Account Details",
	}
	return titles[section]
}

// registerFormatters sets up the section formatters
func (c *MeCommand) registerFormatters() {
	c.formatters["personal"] = formatPersonalInfo
	c.formatters["company"] = formatCompanyInfo
	c.formatters["address"] = formatAddressInfo
	c.formatters["account"] = formatAccountDetails
}

// Section formatters
func formatPersonalInfo(info *api.AccountInfo, section *format.Section) {
	section.AddFields(map[string]string{
		"Name":              fmt.Sprintf("%s %s", info.FirstName, info.Name),
		"Email":             info.Email,
		"Alternative Email": info.SpareEmail,
		"Phone":             formatPhone(info.Phone, info.PhoneCountry),
		"Language":          info.Language,
	})
}

func formatCompanyInfo(info *api.AccountInfo, section *format.Section) {
	section.AddField("Organization", info.Organisation)
	if info.Currency != nil {
		section.AddField(
			"Currency",
			fmt.Sprintf("%s (%s)", info.Currency.Code, info.Currency.Symbol),
		)
	}
}

func formatAddressInfo(info *api.AccountInfo, section *format.Section) {
	section.AddFields(map[string]string{
		"Street":      info.Address,
		"City":        info.City,
		"Postal Code": info.ZIP,
		"Country":     info.Country,
	})
}

func formatAccountDetails(info *api.AccountInfo, section *format.Section) {
	section.AddFields(map[string]string{
		"Customer Code": info.CustomerCode,
		"NIC Handle":    info.NicHandle,
		"Account State": info.State,
		"KYC Validated": fmt.Sprintf("%v", info.KYCValidated),
	})
}

// Helper function to format phone numbers
func formatPhone(phone, country string) string {
	if phone == "" {
		return ""
	}
	if country != "" {
		return fmt.Sprintf("%s (%s)", phone, country)
	}
	return phone
}

