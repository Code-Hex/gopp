package gopp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// ListProxyHandler represents proxy handler for /@v/list
// versionList receieves list of the go release version which is following semantic versioning.
type ListProxyHandler func(w http.ResponseWriter, r *http.Request, versionList []string) error

// AddListProxy registers proxy handler for /@v/list
func (p *Proxy) AddListProxy(h ListProxyHandler) error {
	if h == nil {
		return errors.New("unexpected nil")
	}
	p.versionListHandler = h
	return nil
}

func body2VersionList(body io.Reader) []string {
	ret := make([]string, 0)
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		ret = append(ret, scanner.Text())
	}
	return ret
}

func (p *Proxy) versionListProxy(w http.ResponseWriter, r *http.Request) error {
	// /golang.org/x/net/@v/list
	resp, err := p.request(r.URL.Path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %s", resp.Status)
	}
	vlist := body2VersionList(resp.Body)
	if err := p.versionListHandler(w, r, vlist); err != nil {
		return err
	}
	return nil
}
