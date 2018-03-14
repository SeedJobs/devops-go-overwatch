package abstract

import (
	"fmt"
	"strings"
	"time"

	overwatch "github.com/SeedJobs/devops-go-overwatch"
	"github.com/SeedJobs/devops-go-overwatch/synchro"
	"github.com/SeedJobs/devops-go-overwatch/synchro/git"
)

type Manager struct {
	Expire time.Time
	Conf   *overwatch.IamManagerConfig
	Storer synchro.Store
}

func DefaultManager() *Manager {
	return &Manager{
		Expire: time.Now(),
	}
}

func (m *Manager) Readconfig(conf overwatch.IamManagerConfig) error {
	// Load the configuration needed to interact with Github
	m.Expire = m.Expire.Add(conf.TimeOut)
	synchroType, defined := m.Conf.Additional["Synchro"].(string)
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
		m.Storer = storer
	}
	// As we don't have any data currently stored inside the Manager,
	// Knowning if it had updated is not important
	if _, err := m.Storer.Synced(); err != nil {
		return err
	}
	return nil
}
