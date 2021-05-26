package nomad

import (
	"errors"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"

	"github.com/the-maldridge/dtn/pkg/types"
)

func New() *Nomad {
	return &Nomad{l: hclog.NewNullLogger()}
}

func (n *Nomad) SetParentLogger(l hclog.Logger) {
	n.l = l.Named("nomad")
	n.l.Info("Logging Enabled")
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
	n.c.SetNamespace(namespace)
	qopts := api.QueryOptions{
		Namespace: namespace,
		Prefix:    job,
	}

	jobs, _, err := n.c.Jobs().List(&qopts)
	if err != nil {
		n.l.Warn("Error locating job", "namespace", namespace, "job", job, "error", err)
		return "", err
	}

	for _, j := range jobs {
		if j.Name == job {
			return j.ID, nil
		}
	}
	return "", errors.New("Unknown job")
}

// FindTasksForArtifact looks across all tasks that dtn can see and
// returns a list of tasks that are relevant to the current artifact.
// If you're looking at this its probably because this is too slow and
// you should implement something semi-stateful based on the event
// stream.
func (n *Nomad) FindTasksForArtifact(artifact string) ([]types.NomadTask, error) {
	jobs, _, err := n.c.Jobs().List(&api.QueryOptions{})
	if err != nil {
		return []types.NomadTask{}, err
	}
	tasks := []types.NomadTask{}
	for _, j := range jobs {
		if j.Stop || j.Type == api.JobTypeBatch {
			continue // Don't care about periodic invocations
		}
		job, _, err := n.c.Jobs().Info(j.Name, &api.QueryOptions{Namespace: j.Namespace})
		if err != nil {
			n.l.Warn("Error retrieving job info", "job", j.ID, "error", err)
		}
		for _, group := range job.TaskGroups {
			for _, task := range group.Tasks {
				if v, ok := task.Meta["dtn.enable"]; !ok || v != "enable" {
					continue
				}
				// DTN is enabled for this task.
				n.l.Debug("Found a DTN enabled task",
					"namespace", *job.Namespace,
					"job", *job.ID,
					"group", *group.Name,
					"task", task.Name,
				)

				if v, ok := task.Meta["dns.artifact"]; ok && v == artifact {
					tasks = append(tasks,
						types.NomadTask{
							Namespace: *job.Namespace,
							Job:       *job.ID,
							Group:     *group.Name,
							Task:      task.Name,
						})
				}
			}
		}
	}

	return tasks, nil
}

func (n *Nomad) GetJob(namespace, job string) (*api.Job, error) {
	id, err := n.FindJob(namespace, job)
	if err != nil {
		n.l.Warn("Error getting job", "namespace", namespace, "job", job, "error", err)
		return nil, err
	}
	n.c.SetNamespace(namespace)
	j, _, err := n.c.Jobs().Info(id, nil)
	if err != nil {
		n.l.Warn("Error getting job ID", "namespace", namespace, "job", job, "error", err)
		return nil, err
	}
	return j, nil
}

func (n *Nomad) SetTaskVersion(namespace, job, group, task, version string) error {
	j, err := n.GetJob(namespace, job)
	if err != nil {
		n.l.Warn("Error setting task version", "namespace", namespace, "job", job, "error", err)
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

	n.l.Info("Deploying to namespace", "intent", namespace, "job", j.Namespace)
	n.c.SetNamespace(namespace)
	_, _, err = n.c.Jobs().Register(j, &api.WriteOptions{Namespace: namespace})
	return err
}
