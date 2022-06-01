package utils

import (
	"flag"

	"github.com/spf13/viper"
)

type Config struct {
	Name string
	Env  *EnvConfig
	Flag *FlagConfig
}

type EnvConfig struct {
	Debug          bool
	GRPCPort       int
	GRPCReflection bool
}

type FlagConfig struct {
	HealthCheckMode bool
}

func InitConfig() Config {
	return Config{
		Name: "oauth-poxy",
		Env:  readEnvVars(),
		Flag: readFlags(),
	}
}

func readEnvVars() *EnvConfig {

	// load env vars
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()

	return &EnvConfig{
		Debug:          viper.GetBool("DEBUG"),
		GRPCPort:       viper.GetInt("GRPC_PORT"),
		GRPCReflection: viper.GetBool("GRPC_REFLECTION"),
	}
}

func readFlags() *FlagConfig {

	hcm := flag.Bool("health-check", false, "To run CLI health check mode, default false")
	flag.Parse()

	return &FlagConfig{
		HealthCheckMode: *hcm,
	}
}
