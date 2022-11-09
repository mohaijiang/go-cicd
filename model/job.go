package model

type Job struct {
	Version string           `yaml:"version"`
	Name    string           `yaml:"name"`
	Stages  map[string]Stage `yaml:"stages"`
}
