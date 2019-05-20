package gopp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type InfoProxyHandler func(w http.ResponseWriter, r *http.Request, info *Info) error

func (p *Proxy) AddInfoProxyHandler(h InfoProxyHandler) error {
	if h == nil {
		return errors.New("unexpected nil")
	}
	p.versionInfoHandler = h
	return nil
}

func (p *Proxy) versionInfoProxy(w http.ResponseWriter, r *http.Request, urlPath string) error {
	// /golang.org/x/net/latest
	resp, err := p.request(urlPath)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %s", resp.Status)
	}
	latest, err := body2VersionInfo(resp.Body)
	if err != nil {
		return err
	}
	if err := p.versionInfoHandler(w, r, latest); err != nil {
		return err
	}
	return nil
}

func body2VersionInfo(body io.Reader) (*Info, error) {
	var info Info
	if err := json.NewDecoder(body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}
