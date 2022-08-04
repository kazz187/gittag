package main

import (
	"fmt"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

type Command struct {
	Position *string
	Pre      *string
	Remote   *string
	Repo     *string
	Debug    *bool
}

func NewCommand() *Command {
	app := kingpin.New("gittag", "Semantic versioning tagging tool")
	cmd := &Command{
		Position: app.Flag("position", "the position to increment").Short('p').Enum("major", "minor", "patch"),
		Pre:      app.Flag("pre", "the prerelease suffix").String(),
		Remote:   app.Flag("remote", "the git remote").Default("origin").String(),
		Repo:     app.Flag("repo", "the git repository").Default(".").ExistingDir(),
		Debug:    app.Flag("debug", "enable debug mode").Default("false").Bool(),
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
	g, err := NewGit(*cmd.Repo, *cmd.Remote, auth)
	if err != nil {
		fmt.Println("failed to get git repository:", err)
	}
	tags, err := g.RemoteTags()
	if err != nil {
		fmt.Println("failed to get remote tags:", err)
	}
	sv := NewSemVers(tags, *cmd.Debug)
	latest := sv.Latest()
	fmt.Println("latest:", latest)
	latestPreMap := sv.LatestPre()
	for pre, version := range latestPreMap {
		fmt.Println("latestPre:", pre, version)
	}

	fmt.Println("bump major:", latest.IncMajor())
	fmt.Println("bump minor:", latest.IncMinor())
	fmt.Println("bump patch:", latest.IncPatch())

	for pre, version := range latestPreMap {
		p, n, err := ParsePre(version.Prerelease())
		if err != nil {
			fmt.Println("failed to parse prerelease:", err)
		}
		v, err := version.SetPrerelease(fmt.Sprintf("%s.%d", p, n+1))
		if err != nil {
			fmt.Println("failed to set prerelease:", err)
		}
		fmt.Printf("bump pre(%s): %s\n", pre, v)
	}

	if interactive {

	}
}
