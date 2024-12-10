package commands

import (
    "fmt"
    "ovh-terminal/internal/api"
    "ovh-terminal/internal/format"
)

// MeCommand handles the account info display
type MeCommand struct {
    client *api.Client
}

// NewMeCommand creates a new me command instance
func NewMeCommand(client *api.Client) *MeCommand {
    return &MeCommand{
        client: client,
    }
}

// Execute runs the command
func (c *MeCommand) Execute() (string, error) {
    // Get account info
    info, err := c.client.GetAccountInfo()
    if err != nil {
        return "", err
    }

    // Format the output
    output := format.NewOutputFormatter()

    // Personal Information section
    personal := output.AddSection("Personal Information")
    personal.AddField("Name", fmt.Sprintf("%s %s", info.FirstName, info.Name))
    personal.AddField("Email", info.Email)
    personal.AddField("Alternative Email", info.SpareEmail)
    personal.AddField("Phone", formatPhone(info.Phone, info.PhoneCountry))
    personal.AddField("Language", info.Language)

    // Company Information section
    company := output.AddSection("Company Information")
    company.AddField("Organization", info.Organisation)
    if info.Currency != nil {
        company.AddField("Currency", fmt.Sprintf("%s (%s)", info.Currency.Code, info.Currency.Symbol))
    }

    // Address section
    address := output.AddSection("Address")
    address.AddField("Street", info.Address)
    address.AddField("City", info.City)
    address.AddField("Postal Code", info.ZIP)
    address.AddField("Country", info.Country)

    // Account Details section
    account := output.AddSection("Account Details")
    account.AddField("Customer Code", info.CustomerCode)
    account.AddField("NIC Handle", info.NicHandle)
    account.AddField("Account State", info.State)
    account.AddField("KYC Validated", fmt.Sprintf("%v", info.KYCValidated))

    return output.String(), nil
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