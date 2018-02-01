// Package overwatch is abstraction library that
// enables IAM access control over mutliple providers
// such as GoogleCloudPlatform, Github.
package overwatch

import (
	"errors"
	"time"
)

// ErrNotImplemented Does exactly as its says
var ErrNotImplemented = errors.New("Not yet implemented")

// IamManagerConfig defines the minimal required information
// that a IamManager would require and also works as an expando object.
// Allowing you to store additional information into it.
type IamManagerConfig struct {
	GitLocation string
	TimeOut     time.Duration
	Additional  map[string]interface{}
}

// IamPolicyManager is an abstract to allow
// IAM Configuration management for a given Provider.
// This Provides the base functionality that would be required
// to ensure that IAM management is enforced.
// When an implemented IPolicyManger is created, it is expected that
// it is able to interact with the provider out of the box.
type IamPolicyManager interface {

	// LoadConfiguration loads a configuration object
	// and adds additional configuration to the manager.
	LoadConfiguration(conf IamManagerConfig) error

	// Resource will return all managed resources
	// Ie.
	//   - Users
	//   - Service Accounts / Bots
	//   - Repos
	Resources() []IamResource

	// ListModifiedResources returns the resources
	// that don't match what we currently have stored
	// This is useful for services that want to be notified
	// of changes.
	// Ie.
	// 	 - A Cron that alerts if changes have been made
	//
	// This should only return an error if it was unable
	// to communicate with the provider
	ListModifiedResources() ([]IamResource, error)

	// Resync should apply the stored
	// configuration to the resources/config currently managed.
	// It should return a list of resources that were updated/changed.
	// This should only report an error if the provider is unable to be
	// contacted
	Resync() ([]IamResource, error)
}

// IamResource is a managed item by an IPolicyManger
// These can represent any item that can be configured
// by the provider
type IamResource interface {

	// Name returns the name of the resource
	GetName() string

	// GetType returns the providers resource type
	GetType() string

	// AppliedConfig returns all policies that
	// are currently applied to this resource
	AppliedConfig() []IamConfig
}

// IamConfig is the abstract to represent either a Policy, Role, Rule
type IamConfig interface {

	// Name returns name of the IConfig
	GetName() string

	// String allows object to be correctly read by loggers
	String() string
}
