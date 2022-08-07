package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/Songmu/prompter"
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
	Tag     *string
}

func NewCommand() *Command {
	app := kingpin.New("gittag", "Semantic versioning tagging tool")
	cmd := &Command{
		Segment: app.Flag("segment", "the segment to increment").Short('s').Enum("major", "minor", "patch"),
		Pre:     app.Flag("pre", "the prerelease suffix").String(),
		Remote:  app.Flag("remote", "the git remote").Default("origin").String(),
		Repo:    app.Flag("repo", "the git repository").Default(".").ExistingDir(),
		Debug:   app.Flag("debug", "enable debug mode").Default("false").Bool(),
		Tag:     app.Arg("tag", "the tag to create").String(),
	}
	kingpin.MustParse(app.Parse(os.Args[1:]))
	return cmd
}

func main() {
	interactive := false
	cmd := NewCommand()
	if *cmd.Segment == "" && *cmd.Pre == "" && *cmd.Tag == "" {
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
	latest := sv.Latest
	fmt.Println("latest:", latest)
	for _, pre := range sv.PreRank {
		version, ok := sv.LatestPre[pre]
		if !ok {
			continue
		}
		fmt.Printf("latest pre(%s): %s\n", pre, version)
	}

	v := *cmd.Tag
	if interactive {
		v, err = SelectNextVersion(sv)
		if err != nil {
			fmt.Println("failed to select next version:", err)
			return
		}
	}

	if *cmd.Segment != "" {
		switch *cmd.Segment {
		case "major":
			v = latest.IncMajor().String()
		case "minor":
			v = latest.IncMinor().String()
		case "patch":
			v = latest.IncPatch().String()
		}
	}
	if *cmd.Pre != "" {
		if v == "" {
			version, ok := sv.LatestPre[*cmd.Pre]
			if !ok {
				v = latest.IncPatch().String()
			} else {
				v = version.String()
			}
		}
		verPre := v + "-" + *cmd.Pre
		if version, ok := sv.LatestVerPre[verPre]; !ok {
			v = verPre + ".1"
		} else {
			pre := version.Prerelease()
			_, n, err := ParsePre(pre)
			if err != nil {
				fmt.Println("failed to parse prerelease:", err)
				return
			}
			v = fmt.Sprintf("%s.%d", verPre, n+1)
		}
	}

	if v == "" {
		fmt.Println("no tag specified")
		return
	}
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	fmt.Println("---")
	create := prompter.YN(fmt.Sprintf("Create and push new tag %s to the remote %s?", v, *cmd.Remote), false)
	if create {
		fmt.Printf("Created the tag %s\n", v)
		fmt.Printf("Pushed the tag %s to the remote %s\n", v, *cmd.Remote)
	}
}

func SelectNextVersion(sv *SemVers) (string, error) {
	runewidth.DefaultCondition = &runewidth.Condition{
		EastAsianWidth: false,
	}
	// 次バージョンの候補を表示
	NextVersions := []semver.Version{
		sv.Latest.IncPatch(),
		sv.Latest.IncMinor(),
		sv.Latest.IncMajor(),
	}

	table := tview.NewTable().SetFixed(len(NextVersions), len(sv.PreRank)+2)
	table.SetCell(0, 0, tview.NewTableCell("[red]patch:").SetSelectable(false))
	table.SetCell(1, 0, tview.NewTableCell("[red]minor:").SetSelectable(false))
	table.SetCell(2, 0, tview.NewTableCell("[red]major:").SetSelectable(false))

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
				var err error
				v, err = next.SetPrerelease(fmt.Sprintf("%s.%d", pre, 1))
				if err != nil {
					fmt.Print("failed to set prerelease:", err)
					continue
				}
			}
			table.SetCell(i, j+2, tview.NewTableCell(v.String()))
		}
	}
	table.SetSelectable(true, true).SetBorders(false)
	app := tview.NewApplication()
	var selectedVersion string
	table.SetSelectedFunc(func(row, column int) {
		selectedVersion = table.GetCell(row, column).Text
		app.Stop()
	})
	msg := tview.NewTextView().SetText("Select next version:")
	selected := tview.NewTextView().SetText(table.GetCell(0, 1).Text)
	table.SetSelectionChangedFunc(func(row, column int) {
		selected.SetText(table.GetCell(row, column).Text)
	})
	grid := tview.NewGrid()
	grid.SetRows(1, 0)
	grid.SetColumns(21, 0)
	grid.AddItem(msg, 0, 0, 1, 1, 0, 0, false)
	grid.AddItem(selected, 0, 1, 1, 1, 0, 0, false)
	grid.AddItem(table, 1, 0, 1, 2, 0, 0, true)
	if err := app.SetRoot(grid, true).Run(); err != nil {
		fmt.Println("failed to view bump version select table:", err)
	}
	if selectedVersion == "" {
		return "", errors.New("no version selected")
	}
	return selectedVersion, nil
}
