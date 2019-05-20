package gopp

import (
	"io"
	"net/http"
	"testing"
)

func TestProxy_AddModProxyHandler(t *testing.T) {
	tests := []struct {
		name    string
		h       ModProxyHandler
		wantErr bool
	}{
		{
			name: "Valid",
			h: ModProxyHandler(func(w http.ResponseWriter, r *http.Request, body io.Reader) error {
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
			if err := p.AddModProxyHandler(tt.h); (err != nil) != tt.wantErr {
				t.Errorf("Proxy.AddModProxyHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
