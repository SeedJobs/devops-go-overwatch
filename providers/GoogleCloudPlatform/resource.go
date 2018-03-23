package google

import (
	overwatch "github.com/SeedJobs/devops-go-overwatch"
	yaml "gopkg.in/yaml.v2"
)

type userAccount struct {
	Name  string `json:"Name" yaml:"Name"`
	Email string `json:"Email" yaml:"Email"`
	Type  string `json:"Type" yaml:"Type"`
}

type userCollection []userAccount

func (r userAccount) GetName() string {
	return r.Email
}

func (r userAccount) GetType() string {
	return r.Type
}

func (r userAccount) AppliedConfig() []overwatch.IamConfig {
	return nil
}

func userAccountTransformer(buff []byte) ([]overwatch.IamResource, error) {
	var items userCollection
	if err := yaml.Unmarshal(buff, &items); err != nil {
		return nil, err
	}
	collection := []overwatch.IamResource{}
	for _, obj := range items {
		obj.Type = "ServiceAccount"
		collection = append(collection, obj)
	}
	return collection, nil
}
