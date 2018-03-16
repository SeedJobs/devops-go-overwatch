package github

import (
	"testing"

	overwatch "github.com/SeedJobs/devops-go-overwatch"
)

func TestLoadResourcesFromDisk(t *testing.T) {
	filepath := "./test_data/"
	go func() {
		if r := recover(); r != nil {
			t.Fatal("Recovered in f", r)
		}
	}()
<<<<<<< HEAD
	collection := readFiles(filepath, projectTransformer)
=======
	collection := readFiles(filepath, func(buff []byte) []overwatch.IamResource {
		return nil
	})
>>>>>>> 46ada136ddffe2df60a834fcc98878d89626fd24
	if len(collection) == 0 {
		t.Fatal("Expected data to be read from", filepath)
	}
	for _, item := range collection {
		t.Logf("Got item %+v\n", item)
	}
}
