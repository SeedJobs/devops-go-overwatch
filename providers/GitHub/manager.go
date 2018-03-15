package github

import (
	"context"
	"fmt"
	"net/http"

	overwatch "github.com/SeedJobs/devops-go-overwatch"
	"github.com/SeedJobs/devops-go-overwatch/providers/default"
	gogithub "github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type manager struct {
	organisation string
	client       *gogithub.Client
	store        *abstract.Manager
}

func NewManager() (overwatch.IamPolicyManager, error) {
	return &manager{
		store: abstract.DefaultManager(),
	}, nil
}

func (m *manager) LoadConfiguration(conf overwatch.IamManagerConfig) error {
	// Read all the Manager default configurations
	if err := m.store.Readconfig(conf); err != nil {
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
	return nil
}

func (m *manager) Resources() []overwatch.IamResource {
	return nil
}

func (m *manager) ListModifiedResources() ([]overwatch.IamResource, error) {
	return m.fetchOrgProjects(), nil
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
		repos, resp, err := m.client.Repositories.ListByOrg(context.Background(), m.organisation, opt)
		if err != nil {
			panic(err)
		}
		for _, pro := range repos {
			repo := &project{
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
