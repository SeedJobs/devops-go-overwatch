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
	"time"

	"github.com/SeedJobs/devops-go-synchro"
	"golang.org/x/crypto/ssh"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
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
	store       *synchro.Information
	repo        *git.Repository
	changesMade bool
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
	// When we don't currently have a working copy
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
	// When a local copy can be found
	default:
		// The data does exist
		if g.repo == nil {
			g.repo, err = git.PlainOpen(g.store.Location)
			if err != nil {
				return false, err
			}
		}
		if err = g.downloadContent(); err != nil {
			return false, err
		}
	}
	if err = g.commitContent(); err != nil {
		return false, err
	}
	if err = g.uploadContent(); err != nil {
		return false, err
	}
	return updated || g.changesMade, err
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

func (g *gitSynchro) commitContent() error {
	wrkTree, err := g.repo.Worktree()
	if err != nil {
		return err
	}
	stats, err := wrkTree.Status()
	if err != nil {
		return nil
	}
	var needsCommit bool
	// Going through each file found in the local directory to
	// see if it needs to be added then committed
	for name, status := range stats {
		switch status.Staging {
		case git.Untracked:
			fallthrough
		case git.Modified:
			// Need to stage these changes into git and make a manuall commit
			if _, err = wrkTree.Add(name); err != nil {
				return err
			}
			needsCommit = true
		}
	}
	if needsCommit {
		_, err = wrkTree.Commit("Git Synchro making changes", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Synchro Git",
				Email: "dev@beamery.com",
				When:  time.Now(),
			},
		})
	}
	return err
}

func (g *gitSynchro) uploadContent() error {
	if !g.store.SyncRemote {
		return nil
	}
	opts := &git.PushOptions{}
	if auth, exist := g.store.Additional[authloader].(transport.AuthMethod); exist {
		opts.Auth = auth
	}
	err := g.repo.Push(opts)
	switch err {
	case git.NoErrAlreadyUpToDate:
		err = nil
	}
	return err
}

func (g *gitSynchro) downloadContent() error {
	// Want to pull latest updates for the given branch and report back if we
	// updates any files
	wrkTree, err := g.repo.Worktree()
	if err != nil {
		return err
	}
	opts := &git.PullOptions{}
	if auth, ok := g.store.Additional[authloader]; ok {
		opts.Auth = auth.(transport.AuthMethod)
	}
	err = wrkTree.Pull(opts)
	switch err {
	case nil:
		g.changesMade = true
	case git.NoErrAlreadyUpToDate:
		// No need to report an error if we already have the stuff we need
		err = nil
	}
	return err
}
