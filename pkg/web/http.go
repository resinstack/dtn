package web

import (
	"net/http"

	"github.com/hashicorp/go-hclog"
	"github.com/labstack/echo/v4"
)

func New() *Server {
	x := &Server{
		Echo: echo.New(),
		l:    hclog.NewNullLogger(),
	}

	x.POST("/provider/:provider", x.updateFromProvider)
	x.POST("/update-version/:namespace/:job/:group/:task/:provider", x.updateVersion)
	x.GET("/alive", x.alive)

	return x
}

func (s *Server) SetParentLogger(l hclog.Logger) {
	s.l = l.Named("http")
}

func (s *Server) SetNomadProvider(n Nomad) {
	s.n = n
}

func (s *Server) alive(c echo.Context) error {
	return c.String(http.StatusOK, "not dead yet")
}

func (s *Server) updateVersion(c echo.Context) error {
	n := c.Param("namespace")
	j := c.Param("job")
	g := c.Param("group")
	t := c.Param("task")

	s.l.Debug("Attempting to update task version",
		"namespace", n,
		"job", j,
		"group", g,
		"task", t,
		"provider", c.Param("provider"))

	_, version, err := s.getVersionFromProvider(c)
	if err != nil && err == ErrPing {
		return c.String(http.StatusOK, "pong")
	} else if err != nil {
		return c.String(http.StatusBadRequest, "No version could be extracted")
	}

	if err := s.n.SetTaskVersion(n, j, g, t, version); err != nil {
		s.l.Error("Error updating job", "error", err)
		return c.String(http.StatusInternalServerError, "An internal error has occured")
	}
	return c.String(http.StatusOK, "Task updated successfully")
}

func (s *Server) updateFromProvider(c echo.Context) error {
	artifact, version, err := s.getVersionFromProvider(c)
	if err != nil && err == ErrPing {
		return c.String(http.StatusOK, "pong")
	} else if err != nil {
		return c.String(http.StatusBadRequest, "No version could be extracted")
	}

	tasks, err := s.n.FindTasksForArtifact(artifact)
	if err != nil {
		s.l.Error("Error enumerating tasks", "error", err)
		return c.String(http.StatusInternalServerError, "Could not enumerate tasks for artifact")
	}

	for _, task := range tasks {
		if err := s.n.SetTaskVersion(task.Namespace, task.Job, task.Group, task.Task, version); err != nil {
			s.l.Warn("Failed to update task",
				"namespace", task.Namespace,
				"job", task.Job,
				"group", task.Group,
				"task", task.Task,
				"error", err,
			)
		}
	}

	return c.String(http.StatusOK, "OK")
}

func (s *Server) getVersionFromProvider(c echo.Context) (string, string, error) {
	p := c.Param("provider")
	version := ""

	prvdr, ok := rp[p]
	if !ok {
		s.l.Warn("Request for unknown provider", "request", p, "known", rp)
		return "", "", c.String(http.StatusBadRequest, "Provider is not known")
	}

	version, err := prvdr.ExtractVersion(c.Request())
	if err != nil {
		return "", "", err
	}

	artifact, err := prvdr.ExtractArtifact(c.Request())
	if err != nil {
		return "", "", err
	}
	return artifact, version, nil
}
