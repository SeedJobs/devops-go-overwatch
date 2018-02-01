package google

type config struct {
	Name string `json:"Name" yaml:"Name"`
}

func (c config) GetName() string {
	return c.Name
}

func (c config) String() string {
	return ""
}
