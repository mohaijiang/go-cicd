package model

type Stage struct {
	Steps []Step   `yaml:"steps"`
	Needs []string `yaml:"needs"`
}
