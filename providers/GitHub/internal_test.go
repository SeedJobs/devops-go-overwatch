package github

import (
	"testing"

	"github.com/SeedJobs/devops-go-overwatch/providers/default"
)

func TestLoadResourcesFromDisk(t *testing.T) {
	filepath := "./test_data/"
	collection, err := abstract.ReadFiles(filepath, projectTransformer)
	if err != nil {
		t.Fatal(err)
	}
	if len(collection) == 0 {
		t.Fatal("Expected data to be read from", filepath)
	}
	for _, item := range collection {
		t.Logf("Got item %+v\n", item)
	}
}
