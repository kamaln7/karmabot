package webui

import (
	"fmt"

	"github.com/kamaln7/karmabot/database"
	"github.com/kamaln7/karmabot/ui"

	"github.com/aybabtme/log"
	"github.com/pquerna/otp/totp"
)

// Config contains all the necessary config
// options to start and serve a web UI.
type Config struct {
	ListenAddr, URL, TOTP, FilesPath string
	LeaderboardLimit                 int
	Log                              *log.Log
	Debug                            bool
	DB                               *database.DB
}

// A Provider provides a UI service that can be
// attached to karmabot.
type Provider struct {
	Config *Config
	ui     *UI
}

// ensure that Provider implements the ui.Provider interface
var _ ui.Provider = new(Provider)

// New returns a new instance the web UI provider.
// It also generates a TOTP token and quits if one is not
// passed.
func New(config *Config) (*Provider, error) {
	if config.URL == "" {
		config.URL = fmt.Sprintf("http://%s", config.ListenAddr)
	}

	if config.TOTP == "" {
		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      "karmabot",
			AccountName: "slack",
		})

		if err != nil {
			config.Log.Err(err).Fatal("an error occurred while generating a TOTP key")
		} else {
			config.Log.KV("totpKey", key.Secret()).Fatal("please use the following TOTP key")
		}
	}

	provider := &Provider{
		Config: config,
		ui:     newUI(config),
	}

	return provider, nil
}

// Listen starts the HTTP server.
func (p *Provider) Listen() error {
	p.Config.Log.Info("webui listening")
	p.ui.Listen()

	return nil
}

// GetURL returns the passed URI as a full URL
// with an authentication token that is valid
// for 30 seconds.
func (p *Provider) GetURL(URI string) (string, error) {
	token, err := p.ui.authenticator.GetToken()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s?token=%s", p.Config.URL, URI, token), nil
}
