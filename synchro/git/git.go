// Package git allows to create a git synchro object.
// This enables you to enforce that your local data is synced with your
// remote git repo.
// By default, this project will clone projects to:
// 	synchro/<conf
package git

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"github.com/SeedJobs/devops-go-overwatch/synchro"
	"golang.org/x/crypto/ssh"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

const (
	authloader string = "AuthSigner"
)

// gitSynchro is the Git Synchro implementation
// that allows you to have up to date information
// on remote instances
type gitSynchro struct {
	store *synchro.Information
	repo  *git.Repository
}

// NewSynchro creates a git synchro implementation that
// allows you to interact with a remote git repo
// We want to pass in a copy of the information so that we can modify it
// without changes to the original
func NewSynchro(conf synchro.Information) (synchro.Store, error) {
	if conf.RemoteURL == "" {
		return nil, fmt.Errorf("No remoteURL defined")
	}
	if conf.Branch == "" {
		conf.Branch = "master"
	}
	if conf.Additional == nil {
		conf.Additional = map[string]interface{}{}
	}
	if conf.Location == "" {
		conf.Location = "synchro/" + conf.RemoteURL
	}
	sync := &gitSynchro{
		store: &conf,
	}
	return sync, nil
}

func (g *gitSynchro) Synced() (bool, error) {
	if err := g.configureAuth(); err != nil {
		return false, err
	}
	_, err := os.Stat(g.store.Location)
	var updated bool
	switch {
	case os.IsNotExist(err):
		err = os.MkdirAll(g.store.Location, os.ModePerm)
		if err != nil {
			return false, err
		}
		opts := &git.CloneOptions{
			URL: g.store.RemoteURL,
		}
		if auth, ok := g.store.Additional[authloader]; ok {
			opts.Auth = auth.(transport.AuthMethod)
		}
		// The data does not exist
		g.repo, err = git.PlainClone(g.store.Location, false, opts)
		if err != nil {
			return false, err
		}
		// Setting updated to be true here as we had to clone the directory
		updated = true
	default:
		// The data does exist
		if g.repo == nil {
			g.repo, err = git.PlainOpen(g.store.Location)
			if err != nil {
				return false, err
			}
		}
	}
	// Want to pull latest updates for the given branch and report back if we
	// updates any files
	wrkTree, err := g.repo.Worktree()
	if err != nil {
		return false, err
	}
	opts := &git.PullOptions{}
	if auth, ok := g.store.Additional[authloader]; ok {
		opts.Auth = auth.(transport.AuthMethod)
	}
	err = wrkTree.Pull(opts)
	switch err {
	case git.NoErrAlreadyUpToDate:
		return updated, nil
	case nil:
		return updated, nil
	default:
		return updated, err
	}
}

// GetPath will return the path of which the synced repo is cloned to
func (g *gitSynchro) GetPath() string {
	return g.store.Location
}

// configureAuth will load the auth config loaded from the
// expando object
func (g *gitSynchro) configureAuth() error {
	// check to make sure we don't already have an auth method already configured
	if _, exist := g.store.Additional[authloader]; exist {
		return nil
	}
	tMap, defined := g.store.Additional["auth"]
	if !defined {
		return nil
	}
	store, ok := tMap.(map[string]string)
	if !ok {
		return fmt.Errorf("Unable to convert auth structure to a map of strings to strings")
	}
	switch strings.ToLower(store["Type"]) {
	case "ssh":
		keypath := store["Key-Path"]
		if _, err := os.Stat(keypath); os.IsNotExist(err) {
			return fmt.Errorf("Unable to read default prviate shh key at %s", keypath)
		}
		buf, err := ioutil.ReadFile(keypath)
		if err != nil {
			return err
		}
		signer, err := ssh.ParsePrivateKey([]byte(buf))
		if err != nil {
			return err
		}
		auth := &gitssh.PublicKeys{
			User:   "git",
			Signer: signer,
			HostKeyCallbackHelper: gitssh.HostKeyCallbackHelper{
				HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
					return nil
				},
			},
		}
		g.store.Additional[authloader] = auth
	default:
		return fmt.Errorf("Unknown type trying to be used")
	}
	return nil
}
