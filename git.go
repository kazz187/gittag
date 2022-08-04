package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

type Git struct {
	dir    string
	remote string
	auth   transport.AuthMethod

	r *git.Repository
}

func NewGit(dir, remote string, auth transport.AuthMethod) *Git {
	return &Git{
		dir:    dir,
		remote: remote,
		auth:   auth,
	}
}

func (g *Git) RemoteTags() ([]string, error) {
	repo, err := git.PlainOpen(g.dir)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}
	remote, err := repo.Remote(g.remote)
	if err != nil {
		return nil, fmt.Errorf("failed to get remote: %w", err)
	}
	list, err := remote.List(&git.ListOptions{
		Auth: g.auth,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list remote: %w", err)
	}
	var tags []string
	for _, ref := range list {
		if ref.Name().IsTag() {
			tags = append(tags, ref.Name().Short())
		}

	}
	return tags, nil
}
