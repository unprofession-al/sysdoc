package persistence

import (
	"fmt"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	ssh2 "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"golang.org/x/crypto/ssh"
)

type gitrepo struct {
	repo *git.Repository
	fs   billy.Filesystem
}

func NewGitRepo(keypath, pass, url string) (Persistence, error) {
	gr := &gitrepo{}
	pem, err := os.ReadFile(keypath)
	if err != nil {
		return gr, err
	}
	signer, err := ssh.ParsePrivateKeyWithPassphrase(pem, []byte(pass))
	if err != nil {
		return gr, err
	}
	auth := &ssh2.PublicKeys{User: "git", Signer: signer}
	gr.fs = memfs.New()
	gr.repo, err = git.Clone(memory.NewStorage(), gr.fs, &git.CloneOptions{
		URL:      url,
		Auth:     auth,
		Progress: os.Stdout,
		Mirror:   true,
	})
	return gr, err
}

func (gr *gitrepo) Filesystem() billy.Filesystem {
	return gr.fs
}

func (gr *gitrepo) Branches() ([]string, error) {
	var branches []string
	iter, err := gr.repo.Branches()
	if err != nil {
		return branches, err
	}
	_ = iter.ForEach(func(this *plumbing.Reference) error {
		branches = append(branches, this.Name().String())
		return nil
	})
	return branches, nil
}

func (gr *gitrepo) branch(name string) *plumbing.Reference {
	var ref *plumbing.Reference
	iter, err := gr.repo.Branches()
	if err != nil {
		return ref
	}
	_ = iter.ForEach(func(this *plumbing.Reference) error {
		if this.Name().String() == name {
			ref = this
		}
		return nil
	})
	return ref
}

func (gr *gitrepo) CheckoutBranch(name string) error {
	worktree, err := gr.repo.Worktree()
	if err != nil {
		return err
	}
	branch := gr.branch(name)
	if branch == nil {
		return fmt.Errorf("Branch '%s' does not exist", name)
	}
	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: branch.Name(),
		Force:  true,
	})
	if err != nil {
		return err
	}
	gr.fs = worktree.Filesystem
	return nil
}
