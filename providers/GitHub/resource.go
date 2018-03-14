package github

type project struct {
	Name      string
	Protected []string
	Public    bool
	Teams     map[string]string
}
