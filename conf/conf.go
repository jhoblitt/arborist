package conf

import (
	"io/ioutil"
	"log"
	"strings"

	"gopkg.in/yaml.v3"
)

type ArboristConf struct {
	Repos           []RepoConf `yaml:"repos"`
	ExcludePatterns []string   `yaml:"exclude_patterns"`
	Noop            *bool      `yaml:"noop"`
}

type RepoConf struct {
	FullName string `yaml:"repo"`
	Noop     *bool  `yaml:"noop"`
	Org      string
	Name     string
}

func (s *ArboristConf) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawArboristConf ArboristConf
	if err := unmarshal((*rawArboristConf)(s)); err != nil {
		return err
	}

	if s.Noop == nil {
		v := true
		s.Noop = &v
	}

	return nil
}

func (s *RepoConf) SplitFullName() {
	n := strings.Split(s.FullName, "/")
	s.Org = n[0]
	s.Name = n[1]
}

func (s *RepoConf) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawRepoConf RepoConf
	if err := unmarshal((*rawRepoConf)(s)); err != nil {
		return err
	}

	if s.Noop == nil {
		v := true
		s.Noop = &v
	}

	s.SplitFullName()

	return nil
}

func Parse(conf_file string) ArboristConf {
	raw_conf, err := ioutil.ReadFile(conf_file)
	if err != nil {
		log.Fatal(err)
	}

	var conf ArboristConf
	err = yaml.Unmarshal(raw_conf, &conf)
	if err != nil {
		log.Fatal(err)
	}

	return conf
}
