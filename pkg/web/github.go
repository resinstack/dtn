package web

import (
	"errors"
	"net/http"
	"os"

	"github.com/google/go-github/v35/github"
)

type GitHub struct{}

func init() {
	if rp == nil {
		rp = make(map[string]ReleaseProvider)
	}
	rp["github"] = NewGitHub()
}

func NewGitHub() *GitHub {
	return &GitHub{}
}

func (gh *GitHub) ExtractVersion(r *http.Request) (string, error) {
	payload, err := github.ValidatePayload(r, []byte(os.Getenv("GITHUB_SECRET")))
	if err != nil {
		return "", errors.New("Could not validate webhook")
	}
	event, err := github.ParseWebHook(github.WebHookType(r), payload)

	switch p := event.(type) {
	case github.PackageEvent:
		return *p.GetPackage().GetPackageVersion().Version, nil
	case github.PingEvent:
		return "", ErrPing
	default:
		return "", errors.New("Unknown hook event: "+ p.(*github.Event).GetType())
	}

	return "", errors.New("Could not extract version")
}
