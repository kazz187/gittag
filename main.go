package main

import (
	"fmt"
	"os"

	"github.com/Masterminds/semver/v3"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/mattn/go-runewidth"
	"github.com/rivo/tview"
	"gopkg.in/alecthomas/kingpin.v2"
)

type Command struct {
	Segment *string
	Pre     *string
	Remote  *string
	Repo    *string
	Debug   *bool
}

func NewCommand() *Command {
	app := kingpin.New("gittag", "Semantic versioning tagging tool")
	cmd := &Command{
		Segment: app.Flag("segment", "the segment to increment").Short('s').Enum("major", "minor", "patch"),
		Pre:     app.Flag("pre", "the prerelease suffix").String(),
		Remote:  app.Flag("remote", "the git remote").Default("origin").String(),
		Repo:    app.Flag("repo", "the git repository").Default(".").ExistingDir(),
		Debug:   app.Flag("debug", "enable debug mode").Default("false").Bool(),
	}
	kingpin.MustParse(app.Parse(os.Args[1:]))
	return cmd
}

func main() {
	interactive := false
	cmd := NewCommand()
	if *cmd.Segment == "" && *cmd.Pre == "" {
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
	// 最新バージョンの概要を表示
	for _, pre := range sv.PreRank {
		version, ok := sv.LatestPre[pre]
		if !ok {
			continue
		}
		fmt.Printf("latest pre(%s): %s\n", pre, version)
	}
	latest := sv.Latest
	fmt.Println("latest:", latest)

	if interactive {
		runewidth.DefaultCondition = &runewidth.Condition{
			EastAsianWidth: false,
		}
		// 次バージョンの候補を表示
		NextVersions := []semver.Version{
			latest.IncPatch(),
			latest.IncMinor(),
			latest.IncMajor(),
		}

		table := tview.NewTable().SetFixed(len(NextVersions), len(sv.PreRank)+2)
		table.SetCell(0, 0, tview.NewTableCell("patch").SetSelectable(false))
		table.SetCell(1, 0, tview.NewTableCell("minor").SetSelectable(false))
		table.SetCell(2, 0, tview.NewTableCell("major").SetSelectable(false))

		for i, next := range NextVersions {
			table.SetCell(i, 1, tview.NewTableCell(next.String()))
			for j, pre := range sv.PreRank {
				latestVerPre, _ := sv.LatestVerPre[fmt.Sprintf("%d.%d.%d-%s", next.Major(), next.Minor(), next.Patch(), pre)]

				var v semver.Version
				if latestVerPre != nil &&
					next.Major() == latestVerPre.Major() &&
					next.Minor() == latestVerPre.Minor() &&
					next.Patch() == latestVerPre.Patch() {
					p, n, err := ParsePre(latestVerPre.Prerelease())
					if err != nil {
						fmt.Println("failed to parse prerelease:", err)
						continue
					}
					v, err = latestVerPre.SetPrerelease(fmt.Sprintf("%s.%d", p, n+1))
					if err != nil {
						fmt.Println("failed to set prerelease:", err)
						continue
					}
				} else {
					v, err = next.SetPrerelease(fmt.Sprintf("%s.%d", pre, 1))
					if err != nil {
						fmt.Print("failed to set prerelease:", err)
						continue
					}
				}
				table.SetCell(i, j+2, tview.NewTableCell(v.String()))
			}
			fmt.Println("")
		}
		table.SetSelectable(true, true).SetBorders(true)
		app := tview.NewApplication()
		var selectedVersion string
		table.SetSelectedFunc(func(row, column int) {
			selectedVersion = table.GetCell(row, column).Text
			app.Stop()
		})
		if err := app.SetRoot(table, true).Run(); err != nil {
			fmt.Println("failed to view bump version select table:", err)
		}
		fmt.Printf("selected next version: %s\n", selectedVersion)
	}
}
