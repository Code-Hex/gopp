package gopp

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

const upstreamGoProxyEnv = "UPSTREAM_GOPROXY"

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

// Info defined for json response. see `go help goproxy`
type Info struct {
	Version string    // version string
	Time    time.Time // commit time
}

type Proxy struct {
	u      *url.URL
	client Client

	versionInfoHandler InfoProxyHandler
	versionZipHandler  ZipProxyHandler
	versionModHandler  ModProxyHandler
	versionListHandler ListProxyHandler
}

func NewProxy(c Client) (*Proxy, error) {
	// we expected `export UPSTREAM_GOPROXY=https://original-goproxy.host`
	origHost := os.Getenv(upstreamGoProxyEnv)
	u, err := url.ParseRequestURI(origHost)
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := p.handlers(w, r, r.URL.Path); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func (p *Proxy) handlers(w http.ResponseWriter, r *http.Request, urlPath string) error {
	switch {
	case strings.HasSuffix(urlPath, "/@latest"):
		return p.versionInfoProxy(w, r, urlPath)
	case strings.HasSuffix(urlPath, "/@v/list"):
		return p.versionListProxy(w, r, urlPath)
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
			return p.versionInfoProxy(w, r, urlPath)
		case ".zip":
			return p.versionZipProxy(w, r, urlPath)
		case ".mod":
			return p.versionModProxy(w, r, urlPath)
		}
	}
	return errors.New("unexpected url path")
}
