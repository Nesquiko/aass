package server

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

func LoadConfig[T any](envPrefix string) (*T, error) {
	v := viper.New()

	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg T
	err := v.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("LoadConfig failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
