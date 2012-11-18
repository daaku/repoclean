// Command repoclean cleans old versions of packages from a repo.
package main

import (
	"flag"
	"github.com/daaku/go-alpm"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	repo = flag.String(
		"repo",
		filepath.Join(os.Getenv("HOME"), "pkgs", "repo"),
		"repository directory")
	keep = flag.Uint("keep", 2, "versions to keep")
)

type Arch string

const (
	Any = "any"
	X86 = "x86"
	X64 = "x64"
)

func ParseArch(suffix string) Arch {
	switch suffix {
	case "x86_64.pkg.tar.xz":
		return X64
	}
	return Any
}

type File struct {
	Name    string
	Version string
	Arch    Arch
}

func ParseFile(path string) (*File, error) {
	parts := strings.Split(filepath.Base(path), "-")
	l := len(parts)
	return &File{
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
	files, ok := r.Files[file.Name]
	if ok {
		r.Files[file.Name] = append(files, file)
	} else {
		r.Files[file.Name] = []*File{file}
	}
}

func main() {
	flag.Parse()
	repo, err := ParseRepo(*repo)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(repo)
}
