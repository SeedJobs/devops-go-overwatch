package google

import (
	"testing"

	overwatch "github.com/SeedJobs/devops-go-overwatch"
)

func TestImplementsOverwatch(t *testing.T) {
	var r overwatch.IamResource = userAccount{}
	if _, ok := r.(overwatch.IamResource); !ok {
		t.Fatal("userAccount doesn't implement overwatch.IamResource")
	}
	var c overwatch.IamConfig = config{}
	if _, ok := c.(overwatch.IamConfig); !ok {
		t.Fatal("config does not implement overwatch.IamConfig")
	}
}
