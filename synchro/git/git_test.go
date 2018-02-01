package git_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/SeedJobs/devops-go-overwatch/synchro"
	"github.com/SeedJobs/devops-go-overwatch/synchro/git"
)

func TestImplementsSynchro(t *testing.T) {
	s, err := git.NewSynchro(synchro.Information{
		RemoteURL: "blank",
	})
	if err != nil {
		t.Fatal("Should of been able to create git synchro")
	}
	if _, ok := s.(synchro.Store); !ok {
		t.Fatal("The git synchro does not implement the synchro store")
	}
	if s.GetPath() != "synchro/blank" {
		t.Log("Expected format of the path is: synchro/blank")
		t.Log("Actual resutl returned:", s.GetPath())
		t.Fatal("The default path format is wrong")
	}
}

func TestIncompleteConfig(t *testing.T) {
	_, err := git.NewSynchro(synchro.Information{})
	if err == nil {
		t.Fatal("Not enough information has been provided for the git synchro to sync")
	}
}

func TestPassByCopy(t *testing.T) {
	inf := synchro.Information{
		RemoteURL: "blank",
	}
	man, err := git.NewSynchro(inf)
	if err != nil {
		t.Fatal("Not enough information has been provided for the git synchro to sync")
	}
	if inf.Location == man.GetPath() {
		t.Fatal("Should not equal")
	}
}

func TestCloningCodeViaSSH(t *testing.T) {
	keypath := fmt.Sprintf("%s/.ssh/id_rsa", os.Getenv("HOME"))
	if _, err := os.Stat(keypath); os.IsNotExist(err) {
		t.Skip("Can not find default ssh key to use")
	}
	path := "test/http"
	s, err := git.NewSynchro(synchro.Information{
		RemoteURL: "git@github.com:AlexsJones/kepler.git",
		Location:  path,
		Additional: map[string]interface{}{
			"auth": map[string]string{
				"Type":     "ssh",
				"Key-Path": keypath,
			},
		},
	})
	if err != nil {
		t.Fatal("Failed to create synchro due to", err)
	}
	if s.GetPath() != path {
		t.Fatal("Reporting a different path")
	}
	defer os.RemoveAll("test")
	updated, err := s.Synced()
	if err != nil {
		t.Fatal("Reported issue:", err)
	}
	if !updated {
		t.Fatal("The code should of been updated as the files didn't exist")
	}
	if _, err = os.Stat(path); os.IsNotExist(err) {
		t.Fatal("The folder does not exist")
	}
}

func TestCloningCode(t *testing.T) {
	path := "test/ssh"
	s, err := git.NewSynchro(synchro.Information{
		RemoteURL: "https://github.com/AlexsJones/kepler.git",
		Location:  path,
	})
	if err != nil {
		t.Fatal("Failed to create synchro due to", err)
	}
	if s.GetPath() != path {
		t.Fatal("Reporting a different path")
	}
	defer os.RemoveAll("test")
	updated, err := s.Synced()
	if err != nil {
		t.Fatal("Reported issue:", err)
	}
	if !updated {
		t.Fatal("The code should of been updated as the files didn't exist")
	}
	if _, err = os.Stat(path); os.IsNotExist(err) {
		t.Fatal("The folder does not exist")
	}
}
