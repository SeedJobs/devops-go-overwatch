package google

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	admin "cloud.google.com/go/iam/admin/apiv1"
	overwatch "github.com/SeedJobs/devops-go-overwatch"
	"github.com/SeedJobs/devops-go-synchro"
	"github.com/SeedJobs/devops-go-synchro/git"
	"google.golang.org/api/iterator"
	adminpb "google.golang.org/genproto/googleapis/iam/admin/v1"
	yaml "gopkg.in/yaml.v2"
)

type cloudIamManager struct {
	Project   string
	expire    time.Time
	conf      *overwatch.IamManagerConfig
	storer    synchro.Store
	resources map[string]overwatch.IamResource
}

func NewManager() (overwatch.IamPolicyManager, error) {
	return &cloudIamManager{
		// This will enforce that any operation that depends on expire to happen
		// straight away as no data would have been loaded
		expire:    time.Now(),
		resources: map[string]overwatch.IamResource{},
	}, nil
}

func (m *cloudIamManager) LoadConfiguration(conf overwatch.IamManagerConfig) error {
	m.expire = time.Now().Add(conf.TimeOut)
	project, ok := conf.Additional["Project"].(string)
	if !ok {
		return fmt.Errorf("Unable to convert Project from Additional to a string")
	}
	m.Project = project
	m.conf = &conf
	synchroType, defined := m.conf.Additional["Synchro"].(string)
	if !defined {
		return fmt.Errorf("Unable to find required Synchro listing inside additional map")
	}
	switch strings.ToLower(synchroType) {
	case "git":
		inf := synchro.Information{
			RemoteURL:  conf.GitLocation,
			Additional: conf.Additional,
		}
		if branch, ok := conf.Additional["Branch"].(string); ok {
			inf.Branch = branch
		} else {
			inf.Branch = "master"
		}
		if location, ok := conf.Additional["Location"].(string); ok {
			inf.Location = location
		} else {
			inf.Location = conf.GitLocation
		}
		storer, err := git.NewSynchro(inf)
		if err != nil {
			return err
		}
		m.storer = storer
	}
	// As we don't have any data currently stored inside the Manager,
	// Knowning if it had updated is not important
	if _, err := m.storer.Synced(); err != nil {
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
	if m.storer == nil {
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
	load := func(folder string) (userCollection, error) {
		var folderpath string
		folderpath = path.Join(m.storer.GetPath(), "GoogleCloudPlatform/Project", m.Project, folder)
		if _, err := os.Stat(folderpath); os.IsNotExist(err) {
			return nil, fmt.Errorf("The folder %s does not exist", folderpath)
		}
		filelist, err := ioutil.ReadDir(folderpath)
		if err != nil {
			return nil, err
		}
		var globalitems userCollection
		for _, file := range filelist {
			// Only loading files that have yaml extensions
			matched, _ := regexp.MatchString("^.*\\.([Yy]aml|[Yy]ml)$", file.Name())
			if matched {
				buf, err := ioutil.ReadFile(path.Join(folderpath, file.Name()))
				if err != nil {
					return nil, err
				}
				var items userCollection
				if err = yaml.Unmarshal(buf, &items); err != nil {
					return nil, err
				}
				globalitems = append(globalitems, items...)
			}
		}
		return globalitems, nil
	}
	serviceaccounts, err := load("ServiceAccounts")
	if err != nil {
		return err
	}
	for _, resource := range serviceaccounts {
		resource.Type = "ServiceAccount"
		m.resources[resource.Email] = resource
	}
	// Load Members from disc
	// Load Roles from disc
	return nil
}

func (m *cloudIamManager) update() error {
	if time.Now().After(m.expire) {
		if m.storer == nil {
			return fmt.Errorf("Storer is undefined")
		}
		updated, err := m.storer.Synced()
		switch {
		case err != nil:
			return err
		case updated:
			if err = m.loadFromDisc(); err != nil {
				return err
			}
		}
		// Update the expired time
		m.expire = time.Now().Add(m.conf.TimeOut)
	}
	return nil
}
