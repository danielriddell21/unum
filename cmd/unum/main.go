package main

import (
	"os"

	"github.com/joho/godotenv"

	"github.com/danielriddell21/unum/internal/cli"
)

// version is set at build time via -ldflags "-X main.version=vX.Y.Z"
var version = "dev"

func main() {
	if version == "dev" {
		// In dev builds, load .env for local config overrides. In release builds, ignore .env.
		_ = godotenv.Load() // no-op if .env absent
	}

	if err := cli.Execute(version); err != nil {
		os.Exit(1)
	}
}
