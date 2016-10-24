package blank

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Listen() error {
	return nil
}

func (p *Provider) GetURL(URI string) (string, error) {
	return "", errors.New("webui disabled")
}
