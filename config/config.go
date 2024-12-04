package config

import (
	"fmt"
	"log"
	"stargazer/video-recording/internal/structs"
	"stargazer/video-recording/pkg/env"
)

var App structs.AppConfig

func LoadEnv(env_file string) {
	err := env.Load_env(env_file, &App)
	if err != nil {
		log.Fatalf("failed to load env: %v", err)
	}

	fmt.Printf("ENV's Loaded and validation was success!\n")
}
