package github

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"reflect"

	overwatch "github.com/SeedJobs/devops-go-overwatch"
	"github.com/SeedJobs/devops-go-overwatch/providers/default"
	gogithub "github.com/google/go-github/github"
	"golang.org/x/oauth2"
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
	modified := []overwatch.IamResource{}
	for _, obj := range m.fetchOrgProjects() {
		// early exit if possible
		if _, exist := m.resources[obj.GetType()]; !exist {
			break
		}
		// Check if our stored config matches what is contained
		// inside the current resource
		store, exist := m.resources[obj.GetType()][obj.GetName()]
		switch {
		case !exist:
			// Do nothing
		// Had to use DeepEqual as the standard equal just simply doesn't work in this case
		case !reflect.DeepEqual(store, obj):
			modified = append(modified, obj)
		}
	}
	return modified, nil
}

func (m *manager) Resync() ([]overwatch.IamResource, error) {
	return nil, overwatch.ErrNotImplemented
}

func (m *manager) fetchOrgProjects() []overwatch.IamResource {
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
			panic(err)
		}
		for _, pro := range repos {
			repo := project{
				Name:      pro.GetName(),
				Public:    !pro.GetPrivate(),
				Protected: []string{},
			}
			branches, _, err := m.client.Repositories.ListBranches(context.Background(),
				pro.GetOwner().GetLogin(),
				pro.GetName(),
				&gogithub.ListOptions{})
			if err != nil {
				panic(err)
			}
			for _, branch := range branches {
				if branch.GetProtected() {
					repo.Protected = append(repo.Protected, branch.GetName())
				}
			}
			allRepos = append(allRepos, repo)
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allRepos
}
