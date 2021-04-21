package nomad

import (
	"errors"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"
)

func New() *Nomad {
	return &Nomad{l: hclog.NewNullLogger()}
}

func (n *Nomad) SetParentLogger(l hclog.Logger) {
	n.l = l.Named("nomad")
}

// Connect initializes the nomad client or returns an error.  Nomad's
// standard environment variables will be parsed.
func (n *Nomad) Connect() error {
	c, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return err
	}
	n.c = c
	return nil
}

// FindJob returns a job ID if a job exists, otherwise it returns an
// error that the job can't be found.
func (n *Nomad) FindJob(namespace, job string) (string, error) {
	qopts := api.QueryOptions{
		Namespace: namespace,
		Prefix:    job,
	}

	jobs, _, err := n.c.Jobs().List(&qopts)
	if err != nil {
		return "", err
	}

	for _, j := range jobs {
		if j.Name == job {
			return j.ID, nil
		}
	}
	return "", errors.New("Unknown job")
}

func (n *Nomad) GetJob(namespace, job string) (*api.Job, error) {
	id, err := n.FindJob(namespace, job)
	if err != nil {
		return nil, err
	}
	j, _, err := n.c.Jobs().Info(id, nil)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (n *Nomad) SetTaskVersion(namespace, job, group, task, version string) error {
	j, err := n.GetJob(namespace, job)
	if err != nil {
		return err
	}

	for _, g := range j.TaskGroups {
		if *g.Name != group {
			continue
		}
		for _, t := range g.Tasks {
			if t.Name != task {
				continue
			}

			switch t.Driver {
			case "docker":
				ident := t.Config["image"].(string)
				n.l.Debug("Parsed docker identifier", "identifier", ident)
				image := strings.SplitN(ident, ":", 2)[0]
				newImage := image + ":" + version
				if ident == newImage {
					n.l.Warn("Update has the same image specification; this is not allowed",
						"namespace", namespace,
						"job", job,
						"group", group,
						"task", task,
						"version", version)
					return nil
				}
				n.l.Debug("Patched Reference", "reference", newImage)
				t.Config["image"] = newImage
			}
		}
	}

	_, _, err = n.c.Jobs().Register(j, nil)
	return err
}
