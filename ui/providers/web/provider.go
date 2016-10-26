package web

import (
	"fmt"

	"github.com/aybabtme/log"
	"github.com/pquerna/otp/totp"
)

type Config struct {
	ListenAddr, URL, TOTP, FilesPath string
	LeaderboardLimit                 int
	Log                              *log.Log
	Debug                            bool
}

type Provider struct {
	Config *Config
	ui     *UI
}

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

func (p *Provider) Listen() error {
	p.Config.Log.Info("webui listening")
	p.ui.Listen()

	return nil
}

func (p *Provider) GetURL(URI string) (string, error) {
	token, err := p.ui.authenticator.GetToken()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s?token=%s", p.Config.URL, URI, token), nil
}
