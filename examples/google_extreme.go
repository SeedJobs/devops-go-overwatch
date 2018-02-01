package examples

import (
	"fmt"
	"time"

	overwatch "github.com/SeedJobs/devops-go-overwatch"
	google "github.com/SeedJobs/devops-go-overwatch/providers/GoogleCloudPlatform"
)

/*
   This example will create a GoogleCloudPlatform Iam Manager
   that uses a git synchro to ensure it is up to date with all the
   the data that is listed on the remote server.
*/

func ExampleExtreme() {
	manager, err := google.NewManager()
	if err != nil {
		panic(err)
	}
	err = manager.LoadConfiguration(overwatch.IamManagerConfig{
		GitLocation: "github.com/SeedJobs/test-iam.git",
		// If you want to always ensure that data is up to date,
		// feel free to leave this field blank
		TimeOut: 3 * time.Hour,
		Additional: map[string]interface{}{
			"Branch":   "veted-branch",
			"Location": "/tmp/clone/path",
			"Project":  "my-gcp-project",
			"Synchro":  "git",
			"auth": map[string]string{
				"Type":     "ssh",
				"Key-Path": "/path/to/key",
			},
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Manager loaded resources are:", manager.Resources())
	check, resync := time.Tick(1*time.Second), time.Tick(1*time.Hour)
	for {
		select {
		case <-check:
			modified, err := manager.ListModifiedResources()
			if err != nil {
				panic(err)
			}
			if len(modified) != 0 {
				fmt.Println("Modified resources!!")
				fmt.Println(modified)
			}
		case <-resync:
			modified, err := manager.Resync()
			if err != nil {
				panic(err)
			}
			if len(modified) != 0 {
				fmt.Println("Updated Resources!!")
				fmt.Println(modified)
			}
		}
	}
}
