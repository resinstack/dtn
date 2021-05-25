package web

import (
	"errors"
	"io"
	"net/http"
	"os"
)

type Dumb struct {
	auth string
}

func init() {
	if rp == nil {
		rp = make(map[string]ReleaseProvider)
	}
	rp["dumb"] = NewDumb()
}

func NewDumb() *Dumb {
	return &Dumb{
		auth: os.Getenv("DTN_PRVDR_DUMB_AUTH"),
	}
}

func (d *Dumb) ExtractVersion(r *http.Request) (string, error) {
	// This is really dumb, but it provides a minimum of security
	// and is functionally equivalent to a bearer token without
	// screwing around with checking a bunch of headers.
	user, pass, ok := r.BasicAuth()
	if !ok || user+":"+pass != d.auth {
		return "", errors.New("Unauthorized")
	}

	b, err := io.ReadAll(r.Body)
	return string(b), err
}

func (d *Dumb) ExtractArtifact(r *http.Request) (string, error) {
	return "", errors.New("multi-artifact is not supported for this provider")
}
