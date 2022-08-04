package main

import (
	"fmt"
	"os"

	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"gopkg.in/alecthomas/kingpin.v2"
)

type Command struct {
	Position *string
	Pre      *string
	Remote   *string
	Repo     *string
}

func NewCommand() *Command {
	app := kingpin.New("gittag", "Semantic versioning tagging tool")
	cmd := &Command{
		Position: app.Flag("position", "the position to increment").Short('p').Enum("major", "minor", "patch"),
		Pre:      app.Flag("pre", "the prerelease suffix").String(),
		Remote:   app.Flag("remote", "the git remote").Default("origin").String(),
		Repo:     app.Flag("repo", "the git repository").Default(".").ExistingDir(),
	}
	kingpin.MustParse(app.Parse(os.Args[1:]))
	return cmd
}

func main() {
	interactive := false
	cmd := NewCommand()
	if cmd.Position == nil || cmd.Pre == nil {
		interactive = true
	}
	auth, err := gitssh.DefaultAuthBuilder("git")
	if err != nil {
		fmt.Println("failed to get git auth:", err)
	}

	g := NewGit(*cmd.Repo, *cmd.Remote, auth)
	tags, err := g.RemoteTags()
	if err != nil {
		fmt.Println("failed to get remote tags:", err)
	}
	for i, tag := range tags {
		fmt.Printf("%d: %s\n", i, tag)
	}
	if interactive {

	}
}
