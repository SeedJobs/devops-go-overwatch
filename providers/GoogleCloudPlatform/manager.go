package google

import (
	"context"
	"fmt"
	"path"
	"time"

	admin "cloud.google.com/go/iam/admin/apiv1"
	overwatch "github.com/SeedJobs/devops-go-overwatch"
	"github.com/SeedJobs/devops-go-overwatch/providers/default"
	"google.golang.org/api/iterator"
	adminpb "google.golang.org/genproto/googleapis/iam/admin/v1"
)

type cloudIamManager struct {
	base      *abstract.Manager
	Project   string
	resources map[string]overwatch.IamResource
}

func NewManager() (overwatch.IamPolicyManager, error) {
	return &cloudIamManager{
		// This will enforce that any operation that depends on expire to happen
		// straight away as no data would have been loaded
		base:      abstract.DefaultManager(),
		resources: map[string]overwatch.IamResource{},
	}, nil
}

func (m *cloudIamManager) LoadConfiguration(conf overwatch.IamManagerConfig) error {
	project, ok := conf.Additional["Project"].(string)
	if !ok {
		return fmt.Errorf("Unable to convert Project from Additional to a string")
	}
	m.Project = project
	if err := m.base.Readconfig(conf); err != nil {
		return err
	}
	return m.loadFromDisc()
}

func (m *cloudIamManager) Resources() []overwatch.IamResource {
	m.update()
	res := []overwatch.IamResource{}
	for _, val := range m.resources {
		res = append(res, val)
	}
	return res
}

func (m *cloudIamManager) ListModifiedResources() ([]overwatch.IamResource, error) {
	if m.base.Storer == nil {
		return nil, fmt.Errorf("Unable to load local resources due to misconfigured manager")
	}
	if err := m.update(); err != nil {
		return nil, err
	}
	client, err := m.createClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	req := &adminpb.ListServiceAccountsRequest{
		Name: "projects/" + m.Project,
	}
	it := client.ListServiceAccounts(context.Background(), req)
	modifiedResources := []overwatch.IamResource{}
	// Checking for modified Service accounts
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		serviceAccount := userAccount{
			Name:  resp.GetDisplayName(),
			Email: resp.GetEmail(),
			Type:  "ServiceAccount",
		}
		stored, exist := m.resources[serviceAccount.Email]
		switch {
		case !exist:
			modifiedResources = append(modifiedResources, serviceAccount)
		default:
			if stored != serviceAccount {
				modifiedResources = append(modifiedResources, serviceAccount)
			}
		}
	}
	return modifiedResources, nil
}

func (m *cloudIamManager) Resync() ([]overwatch.IamResource, error) {
	return nil, overwatch.ErrNotImplemented
}

func (m *cloudIamManager) createClient() (*admin.IamClient, error) {
	return admin.NewIamClient(context.Background())
}

func (m *cloudIamManager) loadFromDisc() error {
	// Start from a fresh slate and ensure all previous resources are removed
	for item, _ := range m.resources {
		delete(m.resources, item)
	}
	dir := path.Join(m.base.Storer.GetPath(), "GoogleCloudPlatform/Project", m.Project, "ServiceAcccount")
	serviceaccounts, err := abstract.ReadFiles(dir, userAccountTransformer)
	if err != nil {
		return err
	}
	for _, resource := range serviceaccounts {
		m.resources[resource.GetName()] = resource
	}
	// Load Members from disc
	// Load Roles from disc
	return nil
}

func (m *cloudIamManager) update() error {
	if time.Now().After(m.base.Expire) {
		if m.base.Storer == nil {
			return fmt.Errorf("Storer is undefined")
		}
		updated, err := m.base.Storer.Synced()
		switch {
		case err != nil:
			return err
		case updated:
			if err = m.loadFromDisc(); err != nil {
				return err
			}
		}
		// Update the expired time
		m.base.Expire = time.Now().Add(m.base.Conf.TimeOut)
	}
	return nil
}
