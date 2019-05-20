package gopp

import (
	"net/http"
	"testing"
)

func TestProxy_AddErrHandler(t *testing.T) {
	tests := []struct {
		name    string
		h       ErrHandler
		wantErr bool
	}{
		{
			name: "Valid",
			h: ErrHandler(func(w http.ResponseWriter, r *http.Request, err error) {
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
			if err := p.AddErrHandler(tt.h); (err != nil) != tt.wantErr {
				t.Errorf("Proxy.AddErrHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
