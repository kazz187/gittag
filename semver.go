package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
)

type SemVers struct {
	Versions []*semver.Version
	PreRank  []string
}

func NewSemVers(versions []string, debug bool) *SemVers {
	preCount := map[string]int{}
	sv := SemVers{}
	for _, version := range versions {
		v, err := semver.NewVersion(version)
		if err != nil {
			if debug {
				fmt.Printf("[DEBUG] failed to parse version(%s): %s\n", version, err)
			}
			continue
		}
		sv.Versions = append(sv.Versions, v)

		// Prerelease がある場合は、PreCount をインクリメントする (よく使われている順に Prerelease を並べ替えるため)
		if v.Prerelease() != "" {
			s, _, err := ParsePre(v.Prerelease())
			if err != nil {
				if debug {
					fmt.Printf("[DEBUG] failed to parse prerelease(%s): %s\n", v.Prerelease(), err)
				}
				continue
			}
			i, ok := preCount[s]
			if ok {
				preCount[s] = i + 1
			} else {
				preCount[s] = 1
			}
		}
	}

	// Prerelease をよく使われている順にソートする
	for key := range preCount {
		sv.PreRank = append(sv.PreRank, key)
	}
	sort.Slice(sv.PreRank, func(i, j int) bool {
		return preCount[sv.PreRank[i]] > preCount[sv.PreRank[j]]
	})

	sort.Slice(sv.Versions, func(i, j int) bool {
		return sv.Versions[i].GreaterThan(sv.Versions[j])
	})

	return &sv
}

func (sv *SemVers) Latest() *semver.Version {
	for _, version := range sv.Versions {
		if version.Prerelease() != "" {
			continue
		}
		return version
	}
	return semver.MustParse("v0.0.0")
}

func (sv *SemVers) LatestPre() map[string]*semver.Version {
	preMap := map[string]*semver.Version{}
	for _, version := range sv.Versions {
		if version.Prerelease() == "" {
			continue
		}
		p, _, err := ParsePre(version.Prerelease())
		if err != nil {
			continue
		}
		if _, ok := preMap[p]; ok {
			continue
		}
		preMap[p] = version
	}
	return preMap
}

func ParsePre(str string) (string, int, error) {
	s := strings.Split(str, ".")
	if len(s) != 2 {
		return "", 0, fmt.Errorf("invalid prerelease: %s", str)
	}
	num, err := strconv.Atoi(s[1])
	if err != nil {
		return "", 0, fmt.Errorf("invalid prerelease number: %s", str)
	}
	return s[0], num, nil
}
