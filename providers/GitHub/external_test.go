package github_test

import (
	"fmt"
	"os"
	"testing"

	overwatch "github.com/SeedJobs/devops-go-overwatch"
	github "github.com/SeedJobs/devops-go-overwatch/providers/GitHub"
)

var (
	GitHubEnv     string
	GithubOrg     string
	GitRepository string
)

func init() {
	if GithubOrg = os.Getenv("GITHUB_ORG"); GithubOrg == "" {
		fmt.Println("Unable to determine GITHUB_ORG from env")
		os.Exit(1)
	}
	if GitRepository = os.Getenv("TEST_GIT_URL"); GitRepository == "" {
		fmt.Println("Unable to determine TEST_GIT_URL from env")
		os.Exit(1)
	}
	if GitHubEnv = os.Getenv("GITHUB_TOKEN"); GitHubEnv == "" {
		fmt.Println("Unable to determine GITHUB_TOKEN from env")
		os.Exit(1)
	}
}

func TestInterface(t *testing.T) {
	man, err := github.NewManager()
	if err != nil {
		t.Log("issue:", err)
		t.Fatal("Unable to create manager")
	}
	if _, ok := man.(overwatch.IamPolicyManager); !ok {
		t.Fatal("Does not implement the overwatch IPolicyManager")
	}
}

func TestQuerryingProjects(t *testing.T) {
	man, err := github.NewManager()
	if err != nil {
		t.Log("issue:", err)
		t.Fatal("Unable to create manager")
	}
	keypath := fmt.Sprintf("%s/.ssh/id_rsa", os.Getenv("HOME"))
	clonedir := "test/"
	err = man.LoadConfiguration(overwatch.IamManagerConfig{
		GitLocation: GitRepository,
		Additional: map[string]interface{}{
			"GITHUB_TOKEN": GitHubEnv,
			"GITHUB_ORG":   GithubOrg,
			"Synchro":      "git",
			"Location":     clonedir,
			"auth": map[string]string{
				"Type":     "ssh",
				"Key-Path": keypath,
			},
		},
	})
	defer os.RemoveAll(clonedir)
	if err != nil {
		t.Log("Unable to init Manager")
		t.Fatal(err)
	}
	resources, err := man.ListModifiedResources()
	if err != nil {
		t.Log("Unable to gather modified resources")
		t.Fatal(err)
	}
	for _, resource := range resources {
		t.Logf("\tResource used: %+v", resource)
	}
}
