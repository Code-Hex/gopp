package gopp

import (
	"net/http"
	"testing"
)

func TestProxy_AddListProxy(t *testing.T) {
	tests := []struct {
		name    string
		h       ListProxyHandler
		wantErr bool
	}{
		{
			name: "Valid",
			h: ListProxyHandler(func(w http.ResponseWriter, r *http.Request, l []string) error {
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
			if err := p.AddListProxy(tt.h); (err != nil) != tt.wantErr {
				t.Errorf("Proxy.AddListProxy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
