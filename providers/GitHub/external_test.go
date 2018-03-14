package github_test

import (
	"testing"

	overwatch "github.com/SeedJobs/devops-go-overwatch"
	github "github.com/SeedJobs/devops-go-overwatch/providers/GitHub"
)

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
