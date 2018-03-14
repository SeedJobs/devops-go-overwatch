package github

import overwatch "github.com/SeedJobs/devops-go-overwatch"

type manager struct {
}

func NewManager() (overwatch.IamPolicyManager, error) {
	return &manager{}, overwatch.ErrNotImplemented
}

func (m *manager) LoadConfiguration(conf overwatch.IamManagerConfig) error {
	return overwatch.ErrNotImplemented
}

func (m *manager) Resources() []overwatch.IamResource {
	return nil
}

func (m *manager) ListModifiedResources() ([]overwatch.IamResource, error) {
	return nil, overwatch.ErrNotImplemented
}

func (m *manager) Resync() ([]overwatch.IamResource, error) {
	return nil, overwatch.ErrNotImplemented
}
