package persistence

import (
	"fmt"

	"github.com/go-git/go-billy/v5"
)

type Persistence interface {
	Filesystem() billy.Filesystem
	Branches() ([]string, error)
	CheckoutBranch(string) error
}

type Config struct {
	Filepath string
	Git      struct {
		Keyfile string
		Pass    string
		URL     string
	}
}

func New(c Config) (Persistence, error) {
	if c.Git.URL != "" {
		return NewGitRepo(c.Git.Keyfile, c.Git.Pass, c.Git.URL)
	} else if c.Filepath != "" {
		return NewLocal(c.Filepath)
	}
	return nil, fmt.Errorf("Neither a git url nor a local file path was provided")
}
