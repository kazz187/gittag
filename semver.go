package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
)

type SemVers struct {
	Versions     []*semver.Version
	PreRank      []string
	Latest       *semver.Version
	LatestPre    map[string]*semver.Version
	LatestVerPre map[string]*semver.Version
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

	// すべての Version をソート
	sort.Slice(sv.Versions, func(i, j int) bool {
		return sv.Versions[i].GreaterThan(sv.Versions[j])
	})

	sv.Latest = sv.latest()
	sv.LatestPre, sv.LatestVerPre = sv.latestPre()

	return &sv
}

func (sv *SemVers) latest() *semver.Version {
	for _, version := range sv.Versions {
		if version.Prerelease() != "" {
			continue
		}
		return version
	}
	return semver.MustParse("v0.0.0")
}

func (sv *SemVers) latestPre() (map[string]*semver.Version, map[string]*semver.Version) {
	preMap := map[string]*semver.Version{}    // Key: prerelease (ex: "alpha", "beta", "rc")
	verPreMap := map[string]*semver.Version{} // Key: ver-prerelease (ex: "v1.0.0-alpha", "v1.0.0-beta", "v1.0.0-rc")
	for _, version := range sv.Versions {
		if version.Prerelease() == "" {
			continue
		}
		p, _, err := ParsePre(version.Prerelease())
		if err != nil {
			continue
		}
		if _, ok := preMap[p]; !ok {
			preMap[p] = version
		}

		verPre := fmt.Sprintf("%d.%d.%d-%s", version.Major(), version.Minor(), version.Patch(), p)
		if _, ok := verPreMap[verPre]; !ok {
			verPreMap[verPre] = version
		}
	}
	return preMap, verPreMap
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
