package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func Get(key string) interface{} {
	return viper.Get(key)
}

func GetString(key string) string {
	return viper.GetString(key)
}

func GetBool(key string) bool {
	return viper.GetBool(key)
}

func GetInt(key string) int {
	return viper.GetInt(key)
}

func GetInt64(key string) int64 {
	return viper.GetInt64(key)
}

func GetFloat64(key string) float64 {
	return viper.GetFloat64(key)
}

func GetMap(key string) map[string]interface{} {
	return viper.GetStringMap(key)
}

func GetStringMap(key string) map[string]string {
	return viper.GetStringMapString(key)
}

func GetSlice(key string) []string {
	return viper.GetStringSlice(key)
}

func Init() {
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envVar := "CONFIG"
		jsonConfig := os.Getenv(envVar)
		if jsonConfig == "" {
			panic(fmt.Errorf("environment variable %s not set", envVar))
		}
		var configMap map[string]interface{}
		if err := json.Unmarshal([]byte(jsonConfig), &configMap); err != nil {
			panic(fmt.Errorf("failed to unmarshal JSON from environment variable: %w", err))
		}
		viper.SetConfigType("json")
		if err := viper.MergeConfigMap(configMap); err != nil {
			panic(fmt.Errorf("failed to merge config map into Viper: %w", err))
		}
	} else {
		reader, err := os.Open(fmt.Sprintf("./env.%s.json", envFile))
		if err != nil {
			panic(fmt.Errorf("unable to read config file\n %w", err))
		}
		viper.SetConfigType("json")
		if err := viper.MergeConfig(reader); err != nil {
			panic(fmt.Errorf("failed to merge config map into Viper: %w", err))
		}
	}
}
