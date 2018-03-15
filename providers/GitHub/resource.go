package github

import overwatch "github.com/SeedJobs/devops-go-overwatch"

type project struct {
	Name      string            `json:"Name" yaml:"Name"`
	Protected []string          `json:"Protected" yaml:"Protected"`
	Public    bool              `json:"Public" yaml:"Public"`
	Teams     map[string]string `json:"Teams" yaml:"Public"`
}

func (p project) GetName() string {
	return p.Name
}

func (p project) GetType() string {
	return "Repo"
}

func (p project) AppliedConfig() []overwatch.IamConfig {
	return nil
}
