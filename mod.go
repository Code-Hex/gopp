package gopp

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

// ModProxyHandler represents proxy handler for /@v/v0.0.1.mod
// body receieves mod file. body will close file descripter
// at outside of the handler.
type ModProxyHandler func(w http.ResponseWriter, r *http.Request, body io.Reader) error

// AddModProxyHandler registers proxy handler for /@v/v0.0.1.mod
func (p *Proxy) AddModProxyHandler(h ZipProxyHandler) error {
	if h == nil {
		return errors.New("unexpected nil")
	}
	p.versionZipHandler = h
	return nil
}

func (p *Proxy) versionModProxy(w http.ResponseWriter, r *http.Request, urlPath string) error {
	// /golang.org/x/net/@v/v0.0.1.mod
	resp, err := p.request(urlPath)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %s", resp.Status)
	}
	if err := p.versionModHandler(w, r, resp.Body); err != nil {
		return err
	}
	return nil
}
