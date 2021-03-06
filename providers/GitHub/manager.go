package github

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"reflect"
	"time"

	overwatch "github.com/SeedJobs/devops-go-overwatch"
	"github.com/SeedJobs/devops-go-overwatch/providers/default"
	gogithub "github.com/google/go-github/github"
	"golang.org/x/oauth2"
	yaml "gopkg.in/yaml.v2"
)

type manager struct {
	base         *abstract.Manager
	organisation string
	client       *gogithub.Client
	resources    map[string]map[string]overwatch.IamResource
}

func NewManager() (overwatch.IamPolicyManager, error) {
	return &manager{
		base:      abstract.DefaultManager(),
		resources: map[string]map[string]overwatch.IamResource{},
	}, nil
}

func (m *manager) LoadConfiguration(conf overwatch.IamManagerConfig) error {
	// Read all the Manager default configurations
	if err := m.base.Readconfig(conf); err != nil {
		return err
	}
	// Configure and store resources that are needed for the Client
	var authclient *http.Client = nil
	if token, exist := conf.Additional["GITHUB_TOKEN"].(string); exist {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		authclient = oauth2.NewClient(context.Background(), ts)
	}
	m.client = gogithub.NewClient(authclient)
	if org, exist := conf.Additional["GITHUB_ORG"].(string); exist {
		m.organisation = org
	} else {
		return fmt.Errorf("GITHUB_ORG was not defined in conf additional map")
	}
	return m.readFromDisk()
}

func (m *manager) Resources() []overwatch.IamResource {
	collection := []overwatch.IamResource{}
	for _, dict := range m.resources {
		for _, obj := range dict {
			collection = append(collection, obj)
		}
	}
	return collection
}

// ListModifiedResources will examine resources loaded from Github and check
// them against the expected store configuration.
func (m *manager) ListModifiedResources() ([]overwatch.IamResource, error) {
	collection, err := m.fetchOrgProjects()
	if err != nil {
		return nil, err
	}
	notcached, modified := m.seperateLists(collection)
	return append(notcached, modified...), nil
}

func (m *manager) Resync() ([]overwatch.IamResource, error) {
	items := []overwatch.IamResource{}
	if time.Now().After(m.base.Expire) {
		collection, err := m.fetchOrgProjects()
		if err != nil {
			return nil, err
		}
		notcached, modified := m.seperateLists(collection)
		for _, item := range notcached {
			if _, exist := m.resources[item.GetType()]; !exist {
				m.resources[item.GetType()] = map[string]overwatch.IamResource{}
			}
			m.resources[item.GetType()][item.GetName()] = item
		}
		if err := m.writeToDisk(); err != nil {
			return nil, err
		}
		items = append(items, modified...)
		m.base.Expire = time.Now().Add(m.base.Conf.TimeOut)
	}
	return items, nil
}

func (m *manager) fetchOrgProjects() ([]overwatch.IamResource, error) {
	opt := &gogithub.RepositoryListByOrgOptions{
		ListOptions: gogithub.ListOptions{
			PerPage: 64,
		},
	}
	var allRepos []overwatch.IamResource
	for {
		// This call is limited by the token issuer as it can only see what the issuer can see inside the org
		repos, resp, err := m.client.Repositories.ListByOrg(context.Background(), m.organisation, opt)
		if err != nil {
			return nil, err
		}
		for _, pro := range repos {
			repo := project{
				Name:      pro.GetName(),
				Public:    !pro.GetPrivate(),
				Protected: []string{},
			}
			branchOpts := &gogithub.ListOptions{}
			for {
				branches, branchresp, err := m.client.Repositories.ListBranches(context.Background(),
					pro.GetOwner().GetLogin(),
					pro.GetName(),
					branchOpts)
				if err != nil {
					return nil, err
				}
				for _, branch := range branches {
					if branch.GetProtected() {
						repo.Protected = append(repo.Protected, branch.GetName())
					}
				}
				if branchresp.NextPage == 0 {
					break
				}
				branchOpts.Page = branchresp.NextPage
			}
			allRepos = append(allRepos, repo)
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allRepos, nil
}

func (m *manager) seperateLists(collection []overwatch.IamResource) ([]overwatch.IamResource, []overwatch.IamResource) {
	notcached, modified := []overwatch.IamResource{}, []overwatch.IamResource{}
	for _, item := range collection {
		if _, exist := m.resources[item.GetType()]; !exist {
			notcached = append(notcached, item)
			// early exit on the loop
			continue
		}
		obj, exist := m.resources[item.GetType()][item.GetName()]
		switch {
		case !exist:
			notcached = append(notcached, item)
		case !reflect.DeepEqual(item, obj):
			modified = append(modified, item)
		}
	}
	return notcached, modified
}

func (m *manager) readFromDisk() error {
	// Remove items for the internal cache
	for key, _ := range m.resources {
		delete(m.resources, key)
	}
	// Directory is made up of "path/<Provider>/<Organisation>/<ResourceType>/*.ya?ml"
	dir := path.Join(m.base.Storer.GetPath(), "Github", m.organisation, "Repos")
	loaded, err := abstract.ReadFiles(dir, projectTransformer)
	if err != nil {
		return err
	}
	for _, obj := range loaded {
		if _, ok := m.resources[obj.GetType()]; !ok {
			m.resources[obj.GetType()] = map[string]overwatch.IamResource{}
		}
		m.resources[obj.GetType()][obj.GetName()] = obj
	}
	return nil
}

func (m *manager) writeToDisk() error {
	for key, items := range m.resources {
		dir := path.Join(m.base.Storer.GetPath(), "Github", m.organisation, key+"s")
		data := []overwatch.IamResource{}
		for _, obj := range items {
			data = append(data, obj)
		}
		buff, err := yaml.Marshal(&data)
		if err != nil {
			return err
		}
		f := path.Join(dir, key+".yml")
		// Removing any old file as need just to remove it
		if _, err := os.Stat(f); !os.IsNotExist(err) {
			os.Remove(f)
		}
		// Write the updated file
		if err := ioutil.WriteFile(f, buff, 0644); err != nil {
			return err
		}
	}
	return nil
}
