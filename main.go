package main

import (
	"log"
	"os"

	"github.com/lusingander/ghcv-cli/internal/gh"
	"github.com/lusingander/ghcv-cli/internal/ui"
)

func run(args []string) error {
	cfg, err := gh.LoadConfig()
	if err != nil {
		cfg, err = gh.Authorize()
		if err != nil {
			return err
		}
		if err := gh.SaveConfig(cfg); err != nil {
			return err
		}
	}
	client := gh.NewGitHubClient(cfg)
	return ui.Start(client)
}

func main() {
	if err := run(os.Args); err != nil {
		log.Fatal(err)
	}
}
