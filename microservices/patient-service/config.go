package patientservice

import (
	"fmt"
	"log/slog"
)

type PatientServiceConfig struct {
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

func (c PatientServiceConfig) MongoURI() string {
	return fmt.Sprintf(
		"mongodb://%s:%s@%s:%d/%s?authSource=admin",
		c.Mongo.User,
		c.Mongo.Password,
		c.Mongo.Host,
		c.Mongo.Port,
		c.Mongo.Db,
	)
}
