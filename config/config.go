package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port      string `yaml:"port"`
	ApiGemini string `yaml:"apiGeminiKey"`
	JWTSecret string `yaml:"jwtSecret"`
}

func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	cfg := &Config{}
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
