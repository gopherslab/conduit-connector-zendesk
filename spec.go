package zendesk

import (
	"conduit-connector-zendesk/config"

	sdk "github.com/conduitio/conduit-connector-sdk"
)

func Specification() sdk.Specification {
	return sdk.Specification{
		Name:    "zendesk",
		Summary: "An zendesk source and destination plugin for Conduit, written in Go",
		Version: "v0.1.0",
		Author:  "gopherslab,Inc.",
		DestinationParams: map[string]sdk.Parameter{
			config.ConfigKeyDomain: {
				Default:     "",
				Required:    true,
				Description: "A domain is referred as the organization name to which zendesk is registered",
			},
			config.ConfigKeyUserName: {
				Default:     "",
				Required:    true,
				Description: "Login to zendesk performed using username",
			},
			config.ConfigKeyPassword: {
				Default:     "",
				Required:    true,
				Description: "Password is a string to verify the user for login",
			},
			config.ConfigKeyToken: {
				Default:     "",
				Required:    true,
				Description: "Token is used to OAuth authentication with zendesk client",
			},
		},

		SourceParams: map[string]sdk.Parameter{
			config.ConfigKeyDomain: {
				Default:     "",
				Required:    true,
				Description: "A domain is referred as the organization name to which zendesk is registered",
			},
			config.ConfigKeyUserName: {
				Default:     "",
				Required:    true,
				Description: "Login to zendesk performed using username",
			},
			config.ConfigKeyPassword: {
				Default:     "",
				Required:    true,
				Description: "Password is a string to verify the user for login",
			},
			config.ConfigKeyToken: {
				Default:     "",
				Required:    true,
				Description: "Token is used to OAuth authentication with zendesk client",
			},
		},
	}
}
