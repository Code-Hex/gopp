package gopp

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

// ProxyClient interface represents http client.
// http.Client is satisfied this interface.
type ProxyClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Info defined for json response. see `go help goproxy`
type Info struct {
	Version string    // version string
	Time    time.Time // commit time
}

// Proxy proxies to GOPROXY of upstream.
// this struct is satisfied http.Handler.
type Proxy struct {
	u      *url.URL
	client ProxyClient

	errHandler ErrHandler

	versionInfoHandler InfoProxyHandler
	versionZipHandler  ZipProxyHandler
	versionModHandler  ModProxyHandler
	versionListHandler ListProxyHandler
}

// NewProxy makes proxy of the GOPROXY. returns Proxy struct which is satisfied http.Handler.
func NewProxy(c ProxyClient, upstreamGoProxyHost string) (*Proxy, error) {
	// we expected `upstreamGoProxyHost == "https://original-goproxy.host"`
	u, err := url.ParseRequestURI(upstreamGoProxyHost)
	if err != nil {
		return nil, fmt.Errorf("unexpected host: %v", err)
	}
	return &Proxy{
		u:      u,
		client: c,
	}, nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.makeHandler().ServeHTTP(w, r)
}

func (p *Proxy) request(path string) (*http.Response, error) {
	u := *p.u // clone
	u.Path = path
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	return p.client.Do(req)
}

func (p *Proxy) makeHandler() http.Handler {
	if p.errHandler == nil {
		p.errHandler = defaultErrHandler()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := p.handlers(w, r); err != nil {
			p.errHandler(w, r, err)
		}
	})
}

func (p *Proxy) handlers(w http.ResponseWriter, r *http.Request) error {
	urlPath := r.URL.Path
	switch {
	case strings.HasSuffix(urlPath, "/@latest"):
		return p.versionInfoProxy(w, r)
	case strings.HasSuffix(urlPath, "/@v/list"):
		return p.versionListProxy(w, r)
	default:
		basename := path.Base(urlPath)
		fileExt := filepath.Ext(basename)
		version := strings.TrimSuffix(basename, fileExt)
		// expected semantic version format like v1.0.0
		if !semver.IsValid(version) {
			return fmt.Errorf("unexpected semantic version format: %s", version)
		}
		// expected path like /@v/v0.0.1.info
		if !strings.HasSuffix(urlPath, "/@v/"+basename) {
			return fmt.Errorf("unexpected module path: %s", urlPath)
		}
		switch fileExt {
		case ".info":
			return p.versionInfoProxy(w, r)
		case ".zip":
			return p.versionZipProxy(w, r)
		case ".mod":
			return p.versionModProxy(w, r)
		}
	}
	return errors.New("unexpected url path")
}
