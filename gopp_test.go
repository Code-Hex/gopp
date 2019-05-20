package gopp

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type mockClient struct {
	DoMock func(req *http.Request) (*http.Response, error)
}

func (c *mockClient) Do(req *http.Request) (*http.Response, error) {
	return c.DoMock(req)
}

var (
	versionJSON = func() io.ReadCloser {
		return ioutil.NopCloser(
			strings.NewReader(
				`{"Version":"v0.0.1","Time":"2019-01-02T22:52:24-08:00"}`,
			),
		)
	}
	moduleFILE = func() io.ReadCloser {
		return ioutil.NopCloser(
			strings.NewReader(
				`module github.com/pkg/errors`,
			),
		)
	}
	versionList = func() io.ReadCloser {
		return ioutil.NopCloser(
			strings.NewReader(
				"v0.0.1\nv0.0.2",
			),
		)
	}
	emptyBody = ioutil.NopCloser(nil)
)

func TestNewProxy(t *testing.T) {
	tests := []struct {
		name     string
		upstream string
		wantErr  bool
	}{
		{
			name:     "Valid",
			upstream: "https://localhost",
			wantErr:  false,
		},
		{
			name:     "Invalid",
			upstream: "",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewProxy(http.DefaultClient, tt.upstream)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewProxy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProxy_handlers(t *testing.T) {
	type fields struct {
		client             ProxyClient
		versionInfoHandler InfoProxyHandler
		versionZipHandler  ZipProxyHandler
		versionModHandler  ModProxyHandler
		versionListHandler ListProxyHandler
	}
	tests := []struct {
		name    string
		fields  fields
		urlPath string
		wantErr bool
	}{
		{
			name: "Valid /@latest",
			fields: fields{
				client: &mockClient{
					DoMock: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       versionJSON(),
						}, nil
					},
				},
				versionInfoHandler: InfoProxyHandler(func(w http.ResponseWriter, r *http.Request, info *Info) error {
					if expected := "v0.0.1"; info.Version != expected {
						t.Errorf("expected %s but got %s", expected, info.Version)
					}
					return nil
				}),
			},
			urlPath: "github.com/pkg/errors/@latest",
			wantErr: false,
		},
		{
			name: "Invalid json decode /@latest", // unexpected flow
			fields: fields{
				client: &mockClient{
					DoMock: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(strings.NewReader("")),
						}, nil
					},
				},
			},
			urlPath: "github.com/pkg/errors/@latest",
			wantErr: true,
		},
		{
			name: "Invalid request in /@latest",
			fields: fields{
				client: &mockClient{
					DoMock: func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("error")
					},
				},
			},
			urlPath: "github.com/pkg/errors/@latest",
			wantErr: true,
		},
		{
			name: "Invalid handler of /@latest",
			fields: fields{
				client: &mockClient{
					DoMock: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       versionJSON(),
						}, nil
					},
				},
				versionInfoHandler: InfoProxyHandler(func(w http.ResponseWriter, r *http.Request, info *Info) error {
					return errors.New("error")
				}),
			},
			urlPath: "github.com/pkg/errors/@latest",
			wantErr: true,
		},
		{
			name: "Invalid status code /@latest",
			fields: fields{
				client: &mockClient{
					DoMock: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusBadRequest,
							Body:       emptyBody,
						}, nil
					},
				},
			},
			urlPath: "github.com/pkg/errors/@latest",
			wantErr: true,
		},
		{
			name: "Valid /@v/list",
			fields: fields{
				client: &mockClient{
					DoMock: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       versionList(),
						}, nil
					},
				},
				versionListHandler: ListProxyHandler(func(w http.ResponseWriter, r *http.Request, list []string) error {
					if expected := 2; len(list) != expected {
						t.Errorf("expected %d but got %d", expected, len(list))
					}
					if expected := "v0.0.1"; list[0] != expected {
						t.Errorf("expected %s but got %s", expected, list[0])
					}
					if expected := "v0.0.2"; list[1] != expected {
						t.Errorf("expected %s but got %s", expected, list[1])
					}
					return nil
				}),
			},
			urlPath: "github.com/pkg/errors/@v/list",
			wantErr: false,
		},
		{
			name: "Invalid request in /@v/list",
			fields: fields{
				client: &mockClient{
					DoMock: func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("error")
					},
				},
			},
			urlPath: "github.com/pkg/errors/@v/list",
			wantErr: true,
		},
		{
			name: "Invalid handler of /@v/list",
			fields: fields{
				client: &mockClient{
					DoMock: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       versionJSON(),
						}, nil
					},
				},
				versionListHandler: ListProxyHandler(func(w http.ResponseWriter, r *http.Request, list []string) error {
					return errors.New("error")
				}),
			},
			urlPath: "github.com/pkg/errors/@v/list",
			wantErr: true,
		},
		{
			name: "Invalid status code /@v/list",
			fields: fields{
				client: &mockClient{
					DoMock: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusBadRequest,
							Body:       emptyBody,
						}, nil
					},
				},
			},
			urlPath: "github.com/pkg/errors/@v/list",
			wantErr: true,
		},
		{
			name: "Valid /@v/v0.0.1.info",
			fields: fields{
				client: &mockClient{
					DoMock: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       versionJSON(),
						}, nil
					},
				},
				versionInfoHandler: InfoProxyHandler(func(w http.ResponseWriter, r *http.Request, info *Info) error {
					if expected := "v0.0.1"; info.Version != expected {
						t.Errorf("expected %s but got %s", expected, info.Version)
					}
					return nil
				}),
			},
			urlPath: "github.com/pkg/errors/@v/v0.0.1.info",
			wantErr: false,
		},
		{
			name: "Invalid request in /@v/v0.0.1.info",
			fields: fields{
				client: &mockClient{
					DoMock: func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("error")
					},
				},
			},
			urlPath: "github.com/pkg/errors/@v/v0.0.1.info",
			wantErr: true,
		},
		{
			name: "Invalid handler of /@v/v0.0.1.info",
			fields: fields{
				client: &mockClient{
					DoMock: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       versionJSON(),
						}, nil
					},
				},
				versionInfoHandler: InfoProxyHandler(func(w http.ResponseWriter, r *http.Request, info *Info) error {
					return errors.New("error")
				}),
			},
			urlPath: "github.com/pkg/errors/@v/v0.0.1.info",
			wantErr: true,
		},
		{
			name: "Invalid status code /@v/v0.0.1.info",
			fields: fields{
				client: &mockClient{
					DoMock: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusBadRequest,
							Body:       emptyBody,
						}, nil
					},
				},
			},
			urlPath: "github.com/pkg/errors/@v/v0.0.1.info",
			wantErr: true,
		},
		{
			name:    "Invalid path /@v/hello/v0.0.1.info",
			fields:  fields{},
			urlPath: "github.com/pkg/errors/@v/hello/v0.0.1.info",
			wantErr: true,
		},
		{
			name: "Valid /@v/v0.0.1.zip",
			fields: fields{
				client: &mockClient{
					DoMock: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(nil),
						}, nil
					},
				},
				versionZipHandler: ZipProxyHandler(func(w http.ResponseWriter, r *http.Request, body io.Reader) error {
					return nil
				}),
			},
			urlPath: "github.com/pkg/errors/@v/v0.0.1.zip",
			wantErr: false,
		},
		{
			name: "Invalid request in /@v/v0.0.1.zip",
			fields: fields{
				client: &mockClient{
					DoMock: func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("error")
					},
				},
			},
			urlPath: "github.com/pkg/errors/@v/v0.0.1.zip",
			wantErr: true,
		},
		{
			name: "Invalid handler of /@v/v0.0.1.zip",
			fields: fields{
				client: &mockClient{
					DoMock: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(nil),
						}, nil
					},
				},
				versionZipHandler: ZipProxyHandler(func(w http.ResponseWriter, r *http.Request, body io.Reader) error {
					return errors.New("error")
				}),
			},
			urlPath: "github.com/pkg/errors/@v/v0.0.1.zip",
			wantErr: true,
		},
		{
			name: "Invalid status code /@v/v0.0.1.zip",
			fields: fields{
				client: &mockClient{
					DoMock: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusBadRequest,
							Body:       emptyBody,
						}, nil
					},
				},
			},
			urlPath: "github.com/pkg/errors/@v/v0.0.1.zip",
			wantErr: true,
		},
		{
			name:    "Invalid path /@v/hello/v0.0.1.zip",
			fields:  fields{},
			urlPath: "github.com/pkg/errors/@v/hello/v0.0.1.zip",
			wantErr: true,
		},
		{
			name: "Valid /@v/v0.0.1.mod",
			fields: fields{
				client: &mockClient{
					DoMock: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       moduleFILE(),
						}, nil
					},
				},
				versionModHandler: ModProxyHandler(func(w http.ResponseWriter, r *http.Request, body io.Reader) error {
					buf, err := ioutil.ReadAll(body)
					if err != nil {
						return err
					}
					if expected := "module github.com/pkg/errors"; string(buf) != expected {
						t.Errorf("expected %s but got %s", expected, string(buf))
					}
					return nil
				}),
			},
			urlPath: "github.com/pkg/errors/@v/v0.0.1.mod",
			wantErr: false,
		},
		{
			name: "Invalid request in /@v/v0.0.1.mod",
			fields: fields{
				client: &mockClient{
					DoMock: func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("error")
					},
				},
			},
			urlPath: "github.com/pkg/errors/@v/v0.0.1.mod",
			wantErr: true,
		},
		{
			name: "Invalid handler of /@v/v0.0.1.mod",
			fields: fields{
				client: &mockClient{
					DoMock: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       moduleFILE(),
						}, nil
					},
				},
				versionModHandler: ModProxyHandler(func(w http.ResponseWriter, r *http.Request, body io.Reader) error {
					return errors.New("error")
				}),
			},
			urlPath: "github.com/pkg/errors/@v/v0.0.1.mod",
			wantErr: true,
		},
		{
			name: "Invalid status code /@v/v0.0.1.mod",
			fields: fields{
				client: &mockClient{
					DoMock: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusBadRequest,
							Body:       emptyBody,
						}, nil
					},
				},
			},
			urlPath: "github.com/pkg/errors/@v/v0.0.1.mod",
			wantErr: true,
		},
		{
			name:    "Invalid path /@v/hello/v0.0.1.mod",
			fields:  fields{},
			urlPath: "github.com/pkg/errors/@v/hello/v0.0.1.mod",
			wantErr: true,
		},
		{
			name:    "Invalid version format",
			fields:  fields{},
			urlPath: "github.com/pkg/errors/@v/0.0.1.mod",
			wantErr: true,
		},
		{
			name:    "Invalid path",
			fields:  fields{},
			urlPath: "github.com/pkg/errors/@v/v0.0.1.svg",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Proxy{
				u:                  &url.URL{},
				client:             tt.fields.client,
				versionInfoHandler: tt.fields.versionInfoHandler,
				versionZipHandler:  tt.fields.versionZipHandler,
				versionModHandler:  tt.fields.versionModHandler,
				versionListHandler: tt.fields.versionListHandler,
			}
			req := &http.Request{
				URL: &url.URL{Path: tt.urlPath},
			}
			if err := p.handlers(nil, req); (err != nil) != tt.wantErr {
				t.Errorf("Proxy.handlers() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProxy_ServeHTTP(t *testing.T) {
	type fields struct {
		u                  *url.URL
		client             ProxyClient
		versionInfoHandler InfoProxyHandler
	}
	tests := []struct {
		name     string
		fields   fields
		urlPath  string
		wantCode int
	}{
		{
			name: "Valid /@latest",
			fields: fields{
				client: &mockClient{
					DoMock: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       versionJSON(),
						}, nil
					},
				},
				versionInfoHandler: InfoProxyHandler(func(w http.ResponseWriter, r *http.Request, info *Info) error {
					if expected := "v0.0.1"; info.Version != expected {
						t.Errorf("expected %s but got %s", expected, info.Version)
					}
					return nil
				}),
			},
			urlPath:  "https://localhost/github.com/pkg/errors/@latest",
			wantCode: http.StatusOK,
		},
		{
			name: "Invalid url",
			fields: fields{
				client: &mockClient{
					DoMock: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       versionJSON(),
						}, nil
					},
				},
				versionInfoHandler: InfoProxyHandler(func(w http.ResponseWriter, r *http.Request, info *Info) error {
					if expected := "v0.0.1"; info.Version != expected {
						t.Errorf("expected %s but got %s", expected, info.Version)
					}
					return nil
				}),
			},
			urlPath:  "https://localhost/github.com/pkg/errors/@latest~~~",
			wantCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Proxy{
				u:                  &url.URL{},
				client:             tt.fields.client,
				versionInfoHandler: tt.fields.versionInfoHandler,
			}
			req := httptest.NewRequest("GET", tt.urlPath, nil)
			rec := httptest.NewRecorder()
			p.ServeHTTP(rec, req)
			if rec.Code != tt.wantCode {
				t.Errorf("expected %d but got %d", tt.wantCode, rec.Code)
			}
		})
	}
}

// for 100% coverage
func TestProxy_requestErr(t *testing.T) {
	p := &Proxy{
		u: &url.URL{
			Scheme: "unexpected",
		},
	}
	_, err := p.request("unexpected path")
	if err == nil {
		t.Errorf("unexpected err is nil")
	}
}
