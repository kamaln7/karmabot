package ui

type Provider interface {
	GetURL(URI string) (string, error)
	Listen() error
}
