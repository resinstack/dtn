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
	hook, _ := github.New(github.Options.Secret(os.Getenv("GITHUB_SECRET")))

	payload, err := hook.Parse(r,
		github.PackageEvent,
	)

	switch p := payload.(type) {
	case github.PackageEvent:
		return p.GetPackage().GetPackageVersion().Version, nil
	}

	return "", errors.New("Could not extract version")
}
