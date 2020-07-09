package config

import (
	"time"

	"github.com/spf13/viper"
)

// View is a read-only view of the Open Match configuration.
// New accessors from Viper should be added here.
type View interface {
	IsSet(string) bool
	GetString(string) string
	GetInt(string) int
	GetInt64(string) int64
	GetFloat64(string) float64
	GetStringSlice(string) []string
	GetBool(string) bool
	GetDuration(string) time.Duration
}

// Mutable is a read-write view of the Open Match configuration.
type Mutable interface {
	Set(string, interface{})
	View
}

// Sub returns a subset of configuration filtered by the key.
func Sub(v View, key string) View {
	vcfg, ok := v.(*viper.Viper)
	if ok {
		return vcfg.Sub(key)
	}
	return nil
}
