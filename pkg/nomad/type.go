package nomad

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"
)

// Nomad contains all the wrapped interfaces of the underlying nomad
// client and anything we need to do w.r.t. jobspecs.
type Nomad struct {
	l hclog.Logger
	c *api.Client
}
