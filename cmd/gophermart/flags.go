package main

import (
	"flag"
	"os"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DatabaseUri          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	SecretKey            string `env:"SECRET_KEY"`
}

var ConfigData = Config {}

func parseFlags() {
	flag.StringVar(&ConfigData.RunAddress, "a", "localhost:5050", "server run address")
	flag.StringVar(&ConfigData.DatabaseUri, "d", "", "database uri")
	flag.StringVar(&ConfigData.AccrualSystemAddress, "r", "", "accrual system address")
	flag.StringVar(&ConfigData.SecretKey, "k", "", "secret key for encoding information")

	if v := os.Getenv("RUN_ADDRESS"); v != "" {
		ConfigData.RunAddress = v
	}

	if v := os.Getenv("DATABASE_URI"); v != ""{
		ConfigData.DatabaseUri = v
	}

	if v := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); v != ""{
		ConfigData.AccrualSystemAddress = v
	}

	if v := os.Getenv("SECRET_KEY"); v != ""{
		ConfigData.SecretKey = v
	}
}