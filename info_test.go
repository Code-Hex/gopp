package gopp

import (
	"net/http"
	"testing"
)

func TestProxy_AddInfoProxyHandler(t *testing.T) {
	tests := []struct {
		name    string
		h       InfoProxyHandler
		wantErr bool
	}{
		{
			name: "Valid",
			h: InfoProxyHandler(func(w http.ResponseWriter, r *http.Request, info *Info) error {
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
			if err := p.AddInfoProxyHandler(tt.h); (err != nil) != tt.wantErr {
				t.Errorf("Proxy.AddInfoProxyHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
