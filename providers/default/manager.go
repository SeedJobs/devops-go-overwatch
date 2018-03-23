package abstract

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	overwatch "github.com/SeedJobs/devops-go-overwatch"
	"github.com/SeedJobs/devops-go-synchro"
	"github.com/SeedJobs/devops-go-synchro/git"
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
	m.Conf = &conf
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

<<<<<<< HEAD
func ReadFiles(dir string, transformer func([]byte) []overwatch.IamResource) ([]overwatch.IamResource, error) {
=======
func ReadFiles(dir string, transformer func([]byte) ([]overwatch.IamResource, error)) ([]overwatch.IamResource, error) {
>>>>>>> master
	if f, err := os.Stat(dir); os.IsNotExist(err) && (f != nil && !f.IsDir()) {
		return nil, fmt.Errorf("Unable to load directory %s, %v", dir, err)
	}
	collection := []overwatch.IamResource{}
	filecollection, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, file := range filecollection {
		// Only process files that we expect
		if !regexp.MustCompile("^.*\\.(yaml|yml)$").MatchString(file.Name()) {
			continue
		}
		filepath := path.Join(dir, file.Name())
		buff, err := ioutil.ReadFile(filepath)
		if err != nil {
			return nil, err
		}
<<<<<<< HEAD
		collection = append(collection, transformer(buff)...)
=======
		items, err := transformer(buff)
		if err != nil {
			return nil, err
		}
		collection = append(collection, items...)
>>>>>>> master
	}
	return collection, nil
}
