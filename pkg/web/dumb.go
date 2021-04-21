// +build dev

package web

import (
	"io"
	"net/http"
)

type Dumb struct{}

func init() {
	if rp == nil {
		rp = make(map[string]ReleaseProvider)
	}
	rp["dumb"] = NewDumb()
}

func NewDumb() *Dumb {
	return &Dumb{}
}

func (d *Dumb) ExtractVersion(r *http.Request) (string, error) {
	b, err := io.ReadAll(r.Body)
	return string(b), err
}
