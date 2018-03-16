package github

import (
	overwatch "github.com/SeedJobs/devops-go-overwatch"
	yaml "gopkg.in/yaml.v2"
)

type project struct {
	Name      string   `json:"Name" yaml:"Name"`
	Protected []string `json:"Protected" yaml:"Protected"`
	Public    bool     `json:"Public" yaml:"Public"`
	Teams     []string `json:"Teams" yaml:"Teams"`
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

func projectTransformer(buff []byte) []overwatch.IamResource {
	projects := []project{}
	if err := yaml.Unmarshal(buff, &projects); err != nil {
		panic(err)
	}
	collections := []overwatch.IamResource{}
	for _, pro := range projects {
		collections = append(collections, pro)
	}
	return collections
}
