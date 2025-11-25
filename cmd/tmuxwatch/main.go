// File main.go wires command-line flags and the Bubble Tea program together.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	zone "github.com/alexanderbh/bubblezone/v2"
	tea "github.com/charmbracelet/bubbletea/v2"

	"github.com/steipete/tmuxwatch/internal/tmux"
	"github.com/steipete/tmuxwatch/internal/ui"
)

var version = "0.9.1"

// main configures the tmux client, handles flag modes, and launches Bubble Tea.
func main() {
	zone.NewGlobal()

	var (
		interval   = flag.Duration("interval", time.Second, "tmux poll interval")
		tmuxBin    = flag.String("tmux", "", "path to tmux binary (defaults to PATH lookup)")
		showVer    = flag.Bool("version", false, "print version and exit")
		dump       = flag.Bool("dump", false, "print current tmux snapshot as JSON and exit")
		simulate   = flag.String("debug-click", "", "simulate a mouse left-click at the given coordinates (x,y)")
		traceMouse = flag.Bool("trace-mouse", false, "log mouse hit testing details to stderr")
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

	var debugMsgs []tea.Msg
	if sim := strings.TrimSpace(*simulate); sim != "" {
		parts := strings.Split(sim, ",")
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "invalid --debug-click value %q (want x,y)\n", sim)
			os.Exit(1)
		}
		x, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid debug-click x coordinate %q: %v\n", parts[0], err)
			os.Exit(1)
		}
		y, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid debug-click y coordinate %q: %v\n", parts[1], err)
			os.Exit(1)
		}
		debugMsgs = append(debugMsgs, tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
		if !*traceMouse {
			*traceMouse = true
		}
	}

	model := ui.NewModel(client, *interval, debugMsgs, *traceMouse)
	program := tea.NewProgram(model)

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "tmuxwatch exited with error: %v\n", err)
		os.Exit(1)
	}
}
