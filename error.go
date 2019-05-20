package gopp

import (
	"errors"
	"net/http"
)

// ErrHandler represents handler for handling error
type ErrHandler func(w http.ResponseWriter, r *http.Request, err error)

// AddErrHandler registers error handler
func (p *Proxy) AddErrHandler(h ErrHandler) error {
	if h == nil {
		return errors.New("unexpected nil")
	}
	p.errHandler = h
	return nil
}

func defaultErrHandler() ErrHandler {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
