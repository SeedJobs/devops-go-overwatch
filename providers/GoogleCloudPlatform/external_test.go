package google_test

import (
	"fmt"
	"os"
	"testing"

	overwatch "github.com/SeedJobs/devops-go-overwatch"
	google "github.com/SeedJobs/devops-go-overwatch/providers/GoogleCloudPlatform"
)

var (
	GitRepository string
	GoogleProject string
)

func init() {
	GitRepository = os.Getenv("TEST_GIT_URL")
	GoogleProject = os.Getenv("GOOGLE_PROJECT")
}

func TestHasEnvironmentals(t *testing.T) {
	if GitRepository == "" || GoogleProject == "" {
		t.Log("In order to test the entire extent of the Manager, you'll need to define some variables")
		t.Log("'TEST_GIT_URL' should be the ssh git url used to store the managed resources")
		t.Log("'GOOGLE_PROJECT' should be the name of the google project")
	}
}

func checkError(e error, t *testing.T, method string) {
	switch e {
	case nil:
	case overwatch.ErrNotImplemented:
		t.Fatal("Failed to implement", method)
	}
}

func canAuthWithGCP() bool {
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") != "" {
		return true
	}
	if _, err := os.Stat(fmt.Sprintf("%s/.config/gcloud/application_default_credentials.json", os.Getenv("HOME"))); !os.IsNotExist(err) {
		return true
	}
	return false
}

func TestInterface(t *testing.T) {
	man, err := google.NewManager()
	if err != nil {
		t.Fatal("Unable to create manager")
	}
	if _, ok := man.(overwatch.IamPolicyManager); !ok {
		t.Fatal("Does not implement the overwatch IPolicyManager")
	}
}

func TestAllImplemented(t *testing.T) {
	man, err := google.NewManager()
	if err != nil {
		t.Error("Unable to create manager")
	}
	err = man.LoadConfiguration(overwatch.IamManagerConfig{})
	checkError(err, t, "LoadConfiguration")
	_, err = man.ListModifiedResources()
	checkError(err, t, "ListModification")
	_, err = man.Resync()
	checkError(err, t, "Resync")
}

func TestUninitialiseManager(t *testing.T) {
	man, err := google.NewManager()
	if err != nil {
		t.Error("Unable to create manager")
	}
	if len(man.Resources()) != 0 {
		t.Fatal("Should not have loaded any resources")
	}
}

func TestLoadingResourcesFromVCS(t *testing.T) {
	if GoogleProject == "" || GitRepository == "" {
		t.Skip("Environment variables need to be set")
	}
	man, err := google.NewManager()
	if err != nil {
		t.Error("Unable to create manager")
	}
	clonedir := "test/"
	keypath := fmt.Sprintf("%s/.ssh/id_rsa", os.Getenv("HOME"))
	if _, err := os.Stat(keypath); os.IsNotExist(err) {
		t.Skip("Can not find default ssh key to use")
	}
	err = man.LoadConfiguration(overwatch.IamManagerConfig{
		GitLocation: GitRepository,
		Additional: map[string]interface{}{
			"Location": clonedir,
			"Project":  GoogleProject,
			"Synchro":  "git",
			"auth": map[string]string{
				"Type":     "ssh",
				"Key-Path": keypath,
			},
		},
	})
	if err != nil {
		t.Fatal("Unable to LoadConfigurations due to", err)
	}
	for _, resource := range man.Resources() {
		t.Log("Loaded resource:", resource)
	}
}

func TestListingModifiedResources(t *testing.T) {
	if !canAuthWithGCP() {
		t.Skip("In order to run this test, you must be able to authenticate with GCP")
	}
	if GoogleProject == "" || GitRepository == "" {
		t.Skip("Environment variables need to be set")
	}
	man, err := google.NewManager()
	if err != nil {
		t.Error("Unable to create manager")
	}
	clonedir := "test/"
	keypath := fmt.Sprintf("%s/.ssh/id_rsa", os.Getenv("HOME"))
	if _, err := os.Stat(keypath); os.IsNotExist(err) {
		t.Skip("Can not find default ssh key to use")
	}
	err = man.LoadConfiguration(overwatch.IamManagerConfig{
		GitLocation: GitRepository,
		Additional: map[string]interface{}{
			"Location": clonedir,
			"Project":  GoogleProject,
			"Synchro":  "git",
			"auth": map[string]string{
				"Type":     "ssh",
				"Key-Path": keypath,
			},
		},
	})
	defer os.RemoveAll(clonedir)
	if err != nil {
		t.Fatal("Unable to LoadConfigurations due to", err)
	}
	modified, err := man.ListModifiedResources()
	if err != nil {
		t.Fatal("Manager has returned an error:", err)
	}
	for _, mod := range modified {
		t.Log("Modified resource is:", mod)
	}
}
