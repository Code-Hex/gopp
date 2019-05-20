package gopp

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

type ZipProxyHandler func(w http.ResponseWriter, r *http.Request, body io.Reader) error

func (p *Proxy) AddZipProxyHandler(h ZipProxyHandler) error {
	if h == nil {
		return errors.New("unexpected nil")
	}
	p.versionZipHandler = h
	return nil
}

func (p *Proxy) versionZipProxy(w http.ResponseWriter, r *http.Request, urlPath string) error {
	// /golang.org/x/net/@v/v0.0.1.zip
	resp, err := p.request(urlPath)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %s", resp.Status)
	}
	if err := p.versionZipHandler(w, r, resp.Body); err != nil {
		return err
	}
	return nil
}
