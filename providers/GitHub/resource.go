package github

import overwatch "github.com/SeedJobs/devops-go-overwatch"

type project struct {
	Name      string
	Protected []string
	Public    bool
	Teams     map[string]string
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
