package web

import (
	"github.com/aybabtme/log"
)

type Config struct {
	ListenAddr string
	URL        string
	Log        *log.Logger
}

type Provider struct {
	Config *Config
}

func New(config *Config) *Provider {
	if config.URL == "" {
		config.URL = fmt.Sprintf("http://%s", config.ListenAddr)
	}

	return &Provider{
		Config: config,
	}
}

func (p *Provider) Listen() error {
	p.Config.Log.Info("listening")

	return nil
}

func (p *Provider) GetURL(URI string) (string, error) {
	return fmt.Sprintf("%s%s", p.Config.URL, URI), nil
}
