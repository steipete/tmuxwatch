package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/steipete/tmuxwatch/internal/tmux"
	"github.com/steipete/tmuxwatch/internal/ui"
)

var version = "dev"

func main() {
	var (
		interval = flag.Duration("interval", time.Second, "tmux poll interval")
		tmuxBin  = flag.String("tmux", "", "path to tmux binary (defaults to PATH lookup)")
		showVer  = flag.Bool("version", false, "print version and exit")
		dump     = flag.Bool("dump", false, "print current tmux snapshot as JSON and exit")
	)
	flag.Parse()

	if *showVer {
		fmt.Println("tmuxwatch", version)
		return
	}

	client, err := tmux.NewClient(*tmuxBin)
	// If tmux isn't running, inform the user early.
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to set up tmux client: %v\n", err)
		os.Exit(1)
	}

	if *dump {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		snap, err := client.Snapshot(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to fetch tmux snapshot: %v\n", err)
			os.Exit(1)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(snap); err != nil {
			fmt.Fprintf(os.Stderr, "failed to encode snapshot: %v\n", err)
			os.Exit(1)
		}
		return
	}

	model := ui.NewModel(client, *interval)
	program := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseAllMotion())

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "tmuxwatch exited with error: %v\n", err)
		os.Exit(1)
	}
}
