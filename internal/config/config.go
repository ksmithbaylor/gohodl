package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

var Config config

type config struct {
	TestString string   `mapstructure:"test_string"`
	TestNumber int      `mapstructure:"test_number"`
	TestList   []string `mapstructure:"test_list"`
	Nested     nested   `mapstructure:"nested"`
}

type nested struct {
	NestedString string   `mapstructure:"nested_string"`
	NestedNumber int      `mapstructure:"nested_number"`
	NestedList   []string `mapstructure:"nested_list"`
}

func init() {
	viper.AddConfigPath(".")
	viper.SetConfigFile("config.yml")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(fmt.Errorf("could not read in config: %w", err))
	}

	if err := viper.Unmarshal(&Config); err != nil {
		log.Fatal(fmt.Errorf("invalid config: %w", err))
	}
}
