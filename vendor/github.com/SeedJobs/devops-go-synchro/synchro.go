package synchro

// Information enabled a synchro store to know
// remote VCS it needs to talk to and allows you to dynamically
// add additional information if required as all synchro objects may be different
// in their needs.
type Information struct {
	RemoteURL  string                 `json:"RemoteURL" yaml:"RemoteURL"`
	Branch     string                 `json:"Branch" yaml:"Branch"`
	Location   string                 `json:"Location,omitempty" yaml:"Location,omitempty"`
	Additional map[string]interface{} `json:"Additional,omitempty" yaml:"Additional,omitempty"`
	SyncRemote bool                   `json:"SyncRemote,omitempty" yaml:"SyncRemote,omitempty"`
}

// Store is an interface that allows a VCS to check its
// Currently synced with what is on the remote branch
type Store interface {
	// Synced will force to fetch all remote changes
	// and pull them down locally.
	// Returns true iff it had to update
	// Returns error iff it was unable to talk to remote VCS
	Synced() (bool, error)

	// GetPath returns the path that store cloned to.
	GetPath() string
}
