package google

import overwatch "github.com/SeedJobs/devops-go-overwatch"

type userAccount struct {
	Name  string `json:"Name" yaml:"Name"`
	Email string `json:"Email" yaml:"Email"`
	Type  string `json:"Type" yaml:"Type"`
}

type userCollection []userAccount

func (r userAccount) GetName() string {
	return r.Name
}

func (r userAccount) GetType() string {
	return r.Type
}

func (r userAccount) AppliedConfig() []overwatch.IamConfig {
	return nil
}
