package unifiedio

import "net/url"

type IO interface {
	Connect() error
	SetPath(string) error
	Close() error
}

type UrlPath struct {
	path *url.URL
}

func (i *UrlPath) SetPath(path string) error {

	url, err := url.Parse(path)
	if err != nil {
		return err
	}
	i.path = url
	return nil
}
