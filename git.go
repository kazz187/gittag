package main

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

type Git struct {
	dir    string
	remote string
	auth   transport.AuthMethod

	r *git.Repository
}

func NewGit(dir, remote string, auth transport.AuthMethod) (*Git, error) {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}
	return &Git{
		dir:    dir,
		remote: remote,
		auth:   auth,

		r: repo,
	}, nil
}

func (g *Git) RemoteTags() ([]string, error) {
	remote, err := g.r.Remote(g.remote)
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

func (g *Git) CreateTag(tag string) error {
	l, err := g.r.Log(&git.LogOptions{
		All: true,
	})
	if err != nil {
		return fmt.Errorf("failed to get commit log: %w", err)
	}
	c, err := l.Next()
	if err != nil {
		return fmt.Errorf("failed to get current commit: %w", err)
	}
	_, err = g.r.CreateTag(tag, c.Hash, &git.CreateTagOptions{
		Message: tag,
	})
	return err
}

func (g *Git) PushTag(tag string) error {
	err := g.r.Push(&git.PushOptions{
		Auth:       g.auth,
		RemoteName: g.remote,
		RefSpecs:   []config.RefSpec{config.RefSpec("refs/tags/" + tag + ":refs/tags/" + tag)},
	})
	return err
}
