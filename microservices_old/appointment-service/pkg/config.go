package pkg

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/viper"
)

type AppointmentServiceConfig struct {
	App struct {
		Host     string `mapstructure:"host"`
		Port     string `mapstructure:"port"`
		Timezone string `mapstructure:"timezone"`
	} `mapstructure:"app"`

	Log struct {
		Level slog.Level `mapstructure:"level"`
	} `mapstructure:"log"`

	Mongo struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Db       string `mapstructure:"db"`
	} `mapstructure:"mongo"`
}

func (c AppointmentServiceConfig) MongoURI() string {
	return fmt.Sprintf(
		"mongodb://%s:%s@%s:%d/%s?authSource=admin",
		c.Mongo.User,
		c.Mongo.Password,
		c.Mongo.Host,
		c.Mongo.Port,
		c.Mongo.Db,
	)
}

func LoadConfig() (*AppointmentServiceConfig, error) {
	v := viper.New()

	v.SetEnvPrefix("APPOINTMENTSERVICE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetDefault("app.host", "")
	v.SetDefault("app.port", "")
	v.SetDefault("app.timezone", "")
	v.SetDefault("log.level", "")
	v.SetDefault("mongo.host", "")
	v.SetDefault("mongo.port", "")
	v.SetDefault("mongo.db", "")
	v.SetDefault("mongo.user", "")
	v.SetDefault("mongo.password", "")

	var cfg AppointmentServiceConfig
	err := v.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("LoadConfig failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
