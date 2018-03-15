package github

import (
	"testing"
)

func TestLoadResourcesFromDisk(t *testing.T) {
	filepath := "test_data/users.yaml"
	go func() {
		if r := recover(); r != nil {
			t.Fatal("Recovered in f", r)
		}
	}()
	collection := readFiles(filepath, []project{})
	if len(collection) == 0 {
		t.Fatal("Expected data to be read from", filepath)
	}
	for _, item := range collection {
		t.Logf("Got item %+v\n", item)
	}
}
