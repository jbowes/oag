package config

import (
	"io/ioutil"
	"strings"

	"github.com/go-yaml/yaml"

	"github.com/jbowes/oag/pkg"
)

// Config is the toplevel configuration for running oag
type Config struct {
	Document string `yaml:"document"`
	Output   string `yaml:"output"`
	Package  struct {
		Path string `yaml:"path"`
		Name string `yaml:"name"`
	} `yaml:"package"`

	Boilerplate Boilerplate       `yaml:"boilerplate"`
	Types       map[string]string `yaml:"types"`

	StringFormats map[string]string `yaml:"string_formats"`
}

// Boilerplate defines the options for boilerplate code generation
type Boilerplate struct {
	ClientPrefix string `yaml:"client_prefix"`

	BaseURL  pkg.Visibility `yaml:"base_url"`
	Backend  pkg.Visibility `yaml:"backend"`
	Endpoint pkg.Visibility `yaml:"endpoint"`
}

// Load loads the configuration
func Load(cfgFile string) (*Config, error) {
	b, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return nil, err
	}

	cfg := Config{
		Output: "zz_oag_generated.go",
		Boilerplate: Boilerplate{
			BaseURL:  pkg.Private,
			Backend:  pkg.Public,
			Endpoint: pkg.Private,
		},
	}

	if err = yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	if cfg.Package.Name == "" {
		parts := strings.Split(cfg.Package.Path, "/")
		cfg.Package.Name = parts[len(parts)-1]
	}

	return &cfg, err
}
