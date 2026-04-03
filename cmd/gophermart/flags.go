package main

import (
	"flag"
	"log"
	"os"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	SecretKey            string `env:"SECRET_KEY"`
}

var ConfigData = Config{}

func parseFlags() {
	flag.StringVar(&ConfigData.RunAddress, "a", ":8080", "server run address")
	flag.StringVar(&ConfigData.DatabaseURI, "d", "", "database uri")
	flag.StringVar(&ConfigData.AccrualSystemAddress, "r", "", "accrual system address")
	flag.StringVar(&ConfigData.SecretKey, "k", "", "secret key for encoding information")

	flag.Parse()
	if v := os.Getenv("RUN_ADDRESS"); v != "" {
		ConfigData.RunAddress = v
	}

	if v := os.Getenv("DATABASE_URI"); v != "" {
		ConfigData.DatabaseURI = v
	}

	if v := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); v != "" {
		ConfigData.AccrualSystemAddress = v
	}

	if v := os.Getenv("SECRET_KEY"); v != "" {
		ConfigData.SecretKey = v
	}

	log.Printf(
		"config loaded: run_address=%q database_uri_set=%t accrual_address=%q secret_key_set=%t",
		ConfigData.RunAddress,
		ConfigData.DatabaseURI != "",
		ConfigData.AccrualSystemAddress,
		ConfigData.SecretKey != "",
	)
}
