package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Minecraft MinecraftConfig `yaml:"minecraft"`
	Mods      []Mod           `yaml:"mods"`
	Server    ServerConfig    `yaml:"server"`
}

type MinecraftConfig struct {
	Version          string `yaml:"version"`
	Modloader        string `yaml:"modloader"`
	ModloaderVersion string `yaml:"modloader_version"`
}

type Mod struct {
	Name      string `yaml:"name"`
	Source    string `yaml:"source"`
	ProjectID string `yaml:"project_id"`
	Version   string `yaml:"version"`
	URL       string `yaml:"url"`
	Filename  string `yaml:"filename"`
	Side      string `yaml:"side"` // "both" (default) | "client" | "server"
}

type ServerConfig struct {
	Memory     string `yaml:"memory"`
	MaxPlayers int    `yaml:"max_players"`
	Motd       string `yaml:"motd"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return &cfg, nil
}

// ClientMods returns mods that should be installed on the client (side: both or client).
func (c *Config) ClientMods() []Mod {
	return modsForSide(c.Mods, "client")
}

func modsForSide(mods []Mod, side string) []Mod {
	out := make([]Mod, 0, len(mods))
	for _, m := range mods {
		s := m.Side
		if s == "" {
			s = "both"
		}
		if s == "both" || s == side {
			out = append(out, m)
		}
	}
	return out
}
