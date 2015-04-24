package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

type Package struct {
	Repo    string
	Version string
}

type BunchFile struct {
	Packages    []Package
	DevPackages []Package
	Raw         []string
}

var commentStripRegexp = regexp.MustCompile(`#.*`)
var versionSwapRegexp = regexp.MustCompile(`^(\S+)\s*(\S*)`)

func (b *BunchFile) RawIndex(repo string) (int, bool) {
	for i, packString := range b.Raw {
		parts := strings.Fields(packString)

		if len(parts) < 1 {
			continue
		}

		if parts[0] == repo {
			return i, true
		}
	}

	return 0, false
}

func (b *BunchFile) PackageIndex(repo string) (int, bool) {
	for i, pack := range b.Packages {
		if pack.Repo == repo {
			return i, true
		}
	}

	return 0, false
}

func (b *BunchFile) AddPackage(packString string) error {
	pack := parsePackage(packString)

	index, present := b.RawIndex(pack.Repo)

	if present {
		packIndex, _ := b.PackageIndex(pack.Repo)
		b.Packages[packIndex] = pack

		initialLine := b.Raw[index]

		replacementString := fmt.Sprintf("$1 %s", pack.Version)
		newLine := versionSwapRegexp.ReplaceAllString(initialLine, replacementString)

		b.Raw[index] = newLine
	} else {
		b.Packages = append(b.Packages, pack)
		b.Raw = append(b.Raw, fmt.Sprintf("%s %s", pack.Repo, pack.Version))
	}

	return nil
}

func (b *BunchFile) Save() error {
	err := ioutil.WriteFile("Bunchfile", []byte(strings.Join(b.Raw, "\n")), 0644)

	if err != nil {
		return err
	}

	return nil
}

func readBunchfile() (*BunchFile, error) {
	bunchbytes, err := ioutil.ReadFile("Bunchfile")

	if err != nil {
		return &BunchFile{}, err
	}

	bunch := BunchFile{
		Raw: strings.Split(strings.TrimSpace(string(bunchbytes)), "\n"),
	}

	for _, line := range bunch.Raw {
		line = commentStripRegexp.ReplaceAllLiteralString(line, "")
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		var repo, version string

		packageInfo := strings.Fields(line)

		if len(packageInfo) < 1 {
			continue
		} else {
			repo = packageInfo[0]
		}

		if len(packageInfo) >= 2 {
			version = packageInfo[1]
		}

		bunch.Packages = append(bunch.Packages, Package{Repo: repo, Version: version})
	}

	return &bunch, nil
}
