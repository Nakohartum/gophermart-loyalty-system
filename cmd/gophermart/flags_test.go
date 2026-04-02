package main

import (
	"flag"
	"os"
	"reflect"
	"testing"
)

func TestTextParseFlags(t *testing.T) {
	testCases := []struct {
		name     string
		args     []string
		env      map[string]string
		expected Config
	}{
		{
			name: "defaults",
			args: nil,
			expected: Config{
				RunAddress:           ":8080",
				DatabaseUri:          "",
				AccrualSystemAddress: "",
				SecretKey:            "",
			},
		},
		{
			name: "reads flags",
			args: []string{"-a", ":9090", "-d", "postgres://db", "-r", "http://accrual", "-k", "secret"},
			expected: Config{
				RunAddress:           ":9090",
				DatabaseUri:          "postgres://db",
				AccrualSystemAddress: "http://accrual",
				SecretKey:            "secret",
			},
		},
		{
			name: "env overrides flags",
			args: []string{"-a", ":9090", "-d", "postgres://flag", "-r", "http://flag", "-k", "flag-secret"},
			env: map[string]string{
				"RUN_ADDRESS":            ":7070",
				"DATABASE_URI":           "postgres://env",
				"ACCRUAL_SYSTEM_ADDRESS": "http://env",
				"SECRET_KEY":             "env-secret",
			},
			expected: Config{
				RunAddress:           ":7070",
				DatabaseUri:          "postgres://env",
				AccrualSystemAddress: "http://env",
				SecretKey:            "env-secret",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			oldCommandLine := flag.CommandLine
			oldArgs := os.Args
			oldConfig := ConfigData
			defer func() {
				flag.CommandLine = oldCommandLine
				os.Args = oldArgs
				ConfigData = oldConfig
			}()

			flag.CommandLine = flag.NewFlagSet(tc.name, flag.ContinueOnError)
			ConfigData = Config{}

			for _, key := range []string{"RUN_ADDRESS", "DATABASE_URI", "ACCRUAL_SYSTEM_ADDRESS", "SECRET_KEY"} {
				t.Setenv(key, "")
			}
			for key, value := range tc.env {
				t.Setenv(key, value)
			}

			os.Args = append([]string{oldArgs[0]}, tc.args...)

			parseFlags()

			if !reflect.DeepEqual(ConfigData, tc.expected) {
				t.Fatalf("unexpected config: got %+v want %+v", ConfigData, tc.expected)
			}
		})
	}
}
