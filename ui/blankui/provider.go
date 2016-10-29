package blankui

import (
	"github.com/kamaln7/karmabot/ui"
)

// A Provider provides a UI service that can be
// attached to karmabot.
type Provider struct{}

// ensure that Provider implements the ui.Provider interface
var _ ui.Provider = new(Provider)

// New returns a new blankui Provider instance.
func New() *Provider {
	return &Provider{}
}

// Listen does nothing.
func (p *Provider) Listen() error {
	return nil
}

// GetURL returns an empty string which
// signifies that the UI is disabled.
func (p *Provider) GetURL(URI string) (string, error) {
	return "", nil
}
