package examples

import (
	"fmt"
	"time"

	"github.com/SeedJobs/devops-go-overwatch/synchro"
	"github.com/SeedJobs/devops-go-overwatch/synchro/git"
)

/*
  This example allows to be able to ensure that your local file content
  is updated with the remote's current files.
*/

func ExampleExtreme() {
	s, err := git.NewSynchro(synchro.Information{
		RemoteURL: "github.com/SeedJobs/examples.git",
		Branch:    "veted-branch",
		Location:  "/local/path/to/clone/to",
		Additional: map[string]interface{}{
			"auth": map[string]string{
				"Type":     "ssh",
				"Key-Path": "/path/to/private/ssh/key",
			},
		},
	})
	if err != nil {
		panic(err)
	}
	tick := time.Tick(10 * time.Minute)
	for {
		select {
		case <-tick:
			updated, err := s.Synced()
			if err != nil {
				panic(err)
			}
			if updated {
				fmt.Println("Local files have been updated with remote content")
			}
		}
	}
}
