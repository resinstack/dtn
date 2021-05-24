package web

import (
	"net/http"

	"github.com/hashicorp/go-hclog"
	"github.com/labstack/echo/v4"

	"github.com/the-maldridge/dtn/pkg/types"
)

var (
	rp map[string]ReleaseProvider
)

type Server struct {
	*echo.Echo

	l hclog.Logger
	n Nomad
}

type Nomad interface {
	SetTaskVersion(string, string, string, string, string) error
	FindTasksForArtifact(string) ([]types.NomadTask, error)
}

type ReleaseProvider interface {
	ExtractVersion(*http.Request) (string, error)
	ExtractArtifact(*http.Request) (string, error)
}
