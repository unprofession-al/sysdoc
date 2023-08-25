package persistence

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
)

type local struct {
	fs billy.Filesystem
}

func NewLocal(path string) (Persistence, error) {
	l := &local{
		fs: osfs.New(path),
	}
	return l, nil
}

func (l *local) Filesystem() billy.Filesystem {
	return l.fs
}

func (l *local) Branches() ([]string, error) {
	return []string{"local"}, nil
}

func (l *local) CheckoutBranch(name string) error {
	return nil
}
