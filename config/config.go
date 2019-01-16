package config

import (
	"github.com/spf13/viper"
)

func init() {
	viperConfig.AutomaticEnv()
}

const (
	// DefaultRetryTimes default value
	DefaultRetryTimes = 3
	// DefaultBufferLen default value
	DefaultBufferLen = 100
	// DefaultPort default value
	DefaultPort = 19999
	// DefaultOutParallel default value
	DefaultOutParallel = 10
)

var (
	viperConfig = viper.New()
)

// GetViper get viper
func GetViper() *viper.Viper {
	return viperConfig
}

// GetInt get int value
func GetInt(key string) int {
	return viperConfig.GetInt(key)
}
