package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Site struct {
	Token  string `yaml:"token"`
	APIURL string `yaml:"api_url"`
}

type GlobalConfig struct {
	Sites map[string]*Site `yaml:"sites"`
}

type DirectoryConfig struct {
	Site   string `yaml:"site,omitempty"`
	Token  string `yaml:"token,omitempty"`
	APIURL string `yaml:"api_url,omitempty"`
}

func DefaultGlobalConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "stqry", "config.yaml")
}

func LoadGlobalConfig(path string) (*GlobalConfig, error) {
	cfg := &GlobalConfig{Sites: make(map[string]*Site)}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if cfg.Sites == nil {
		cfg.Sites = make(map[string]*Site)
	}
	return cfg, nil
}

func SaveGlobalConfig(cfg *GlobalConfig, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	return os.WriteFile(path, data, 0600)
}

func FindDirectoryConfig(startDir string) (*DirectoryConfig, error) {
	dir := startDir
	for {
		for _, name := range []string{"stqry.yaml", "stqry.yml"} {
			candidate := filepath.Join(dir, name)
			data, err := os.ReadFile(candidate)
			if err == nil {
				var cfg DirectoryConfig
				if err := yaml.Unmarshal(data, &cfg); err != nil {
					return nil, fmt.Errorf("parsing %s: %w", candidate, err)
				}
				return &cfg, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return &DirectoryConfig{}, nil
		}
		dir = parent
	}
}

func SaveDirectoryConfig(dir string, cfg *DirectoryConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling dir config: %w", err)
	}
	return os.WriteFile(filepath.Join(dir, "stqry.yaml"), data, 0600)
}

func ResolveSite(global *GlobalConfig, flagSite string, dirCfg *DirectoryConfig) (*Site, error) {
	// --site flag takes priority: look up in global config.
	if flagSite != "" {
		site, ok := global.Sites[flagSite]
		if !ok {
			return nil, fmt.Errorf("site %q not found in config. Run `stqry config add-site --name=%s --token=<token> --api-url=<url>`", flagSite, flagSite)
		}
		return site, nil
	}

	// STQRY_SITE environment variable.
	if envSite := os.Getenv("STQRY_SITE"); envSite != "" {
		site, ok := global.Sites[envSite]
		if !ok {
			return nil, fmt.Errorf("site %q (from STQRY_SITE) not found in config", envSite)
		}
		return site, nil
	}

	// Directory config with inline credentials takes next priority.
	if dirCfg != nil && dirCfg.Token != "" && dirCfg.APIURL != "" {
		return &Site{Token: dirCfg.Token, APIURL: dirCfg.APIURL}, nil
	}

	// Directory config referencing a named global site.
	if dirCfg != nil && dirCfg.Site != "" {
		site, ok := global.Sites[dirCfg.Site]
		if !ok {
			return nil, fmt.Errorf("site %q not found in config. Run `stqry config add-site --name=%s --token=<token> --api-url=<url>`", dirCfg.Site, dirCfg.Site)
		}
		return site, nil
	}

	return nil, fmt.Errorf("no site specified. Use --site or run `stqry config init` in this directory")
}
