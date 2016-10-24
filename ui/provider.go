package ui

type Provider interface {
	GetURL(URI string) string
	Listen()
}
