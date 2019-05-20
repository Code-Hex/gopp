package gopp

import (
	"io"
	"net/http"
	"testing"
)

func TestProxy_AddZipProxyHandler(t *testing.T) {
	tests := []struct {
		name    string
		h       ZipProxyHandler
		wantErr bool
	}{
		{
			name: "Valid",
			h: ZipProxyHandler(func(w http.ResponseWriter, r *http.Request, body io.Reader) error {
				return nil
			}),
			wantErr: false,
		},
		{
			name:    "Invalid",
			h:       nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Proxy{}
			if err := p.AddZipProxyHandler(tt.h); (err != nil) != tt.wantErr {
				t.Errorf("Proxy.AddZipProxyHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
