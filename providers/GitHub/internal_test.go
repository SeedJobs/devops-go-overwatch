package github

import (
	"testing"
)

func TestLoadResourcesFromDisk(t *testing.T) {
	filepath := "./test_data/"
	collection, err := readFiles(filepath, projectTransformer)
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
