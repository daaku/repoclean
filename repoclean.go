// Command repoclean cleans old versions of packages from a repo.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Jguer/go-alpm"
)

var (
	repo = flag.String("repo", "/var/lib/pkgrepo", "repository directory")
	keep = flag.Int("keep", 1, "versions to keep")
)

type Arch string

const (
	Any   = "any"
	X86   = "x86"
	X64   = "x64"
	Arm6H = "armv6h"
	Arm7H = "armv7h"
)

func ParseArch(suffix string) Arch {
	switch suffix {
	case "x86_64.pkg.tar.xz":
		return X64
	case "armv6h.pkg.tar.xz":
		return Arm6H
	case "armv7h.pkg.tar.xz":
		return Arm7H
	case "any.pkg.tar.xz":
		return Any
	}
	panic(fmt.Sprintf("unknown arch: %s", suffix))
}

type File struct {
	Path    string
	Name    string
	Version string
	Arch    Arch
}

func ParseFile(path string) (*File, error) {
	parts := strings.Split(filepath.Base(path), "-")
	l := len(parts)
	return &File{
		Path:    path,
		Name:    strings.Join(parts[0:l-3], "-"),
		Version: strings.Join(parts[l-3:l-1], "-"),
		Arch:    ParseArch(parts[l-1]),
	}, nil
}

func (f *File) String() string {
	return strings.Join([]string{f.Name, f.Version, string(f.Arch)}, "-")
}

type Files []*File

func (f Files) Len() int      { return len(f) }
func (f Files) Swap(i, j int) { f[i], f[j] = f[j], f[i] }

type ByVersion struct{ Files }

func (f ByVersion) Less(i, j int) bool {
	if alpm.VerCmp(f.Files[i].Version, f.Files[j].Version) == 1 {
		return true
	}
	return false
}

type Repo struct {
	Files map[string][]*File
}

func ParseRepo(path string) (*Repo, error) {
	repo := &Repo{Files: make(map[string][]*File)}
	err := filepath.Walk(
		path,
		func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			if filepath.Ext(path) != ".xz" {
				return nil
			}
			file, err := ParseFile(path)
			if err != nil {
				return err
			}
			repo.Add(file)
			return nil
		})
	if err != nil {
		return nil, err
	}
	for _, files := range repo.Files {
		sort.Sort(ByVersion{files})
	}
	return repo, nil
}

func (r *Repo) Add(file *File) {
	name := file.Name + string(file.Arch)
	files, ok := r.Files[name]
	if ok {
		r.Files[name] = append(files, file)
	} else {
		r.Files[name] = []*File{file}
	}
}

func (r *Repo) Prune(keep int) error {
	for _, files := range r.Files {
		if len(files) > keep {
			for _, file := range files[keep:len(files)] {
				err := os.Remove(file.Path)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func main() {
	flag.Parse()
	repo, err := ParseRepo(*repo)
	if err != nil {
		log.Fatal(err)
	}
	repo.Prune(*keep)
}
