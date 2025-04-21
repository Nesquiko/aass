package pkg

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/viper"
)

type MedicalHistoryServiceConfig struct {
	App struct {
		Host     string `mapstructure:"host"`
		Port     string `mapstructure:"port"`
		Timezone string `mapstructure:"timezone"`
	} `mapstructure:"app"`

	Log struct {
		Level slog.Level `mapstructure:"level"`
	} `mapstructure:"log"`
}

func LoadConfig() (*MedicalHistoryServiceConfig, error) {
	v := viper.New()

	v.SetEnvPrefix("MEDICALHISTORYSERVICE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetDefault("app.host", "")
	v.SetDefault("app.port", "")
	v.SetDefault("app.timezone", "")
	v.SetDefault("log.level", "")

	var cfg MedicalHistoryServiceConfig
	err := v.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("LoadConfig failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
