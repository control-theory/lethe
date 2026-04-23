package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"time"

	"github.com/control-theory/lethe/internal/config"
	"github.com/control-theory/lethe/internal/dashboard"
	"github.com/control-theory/lethe/internal/graph"
	mcpserver "github.com/control-theory/lethe/internal/mcp"
	"github.com/control-theory/lethe/internal/singleton"
)

var version = "dev"

func main() {
	stdio := flag.Bool("stdio", false, "Run MCP server over stdin/stdout")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("lethe", version)
		os.Exit(0)
	}

	// All logging goes to stderr — stdout is the MCP transport.
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// Subcommand: dashboard
	if flag.NArg() > 0 && flag.Arg(0) == "dashboard" {
		openDashboard()
		return
	}

	if *stdio {
		runStdio()
		return
	}

	fmt.Fprintln(os.Stderr, "usage: lethe --stdio | lethe dashboard")
	os.Exit(1)
}

func runStdio() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Open database
	dbPath, err := config.DBPath()
	if err != nil {
		slog.Error("failed to resolve database path", "error", err)
		os.Exit(1)
	}
	db, err := graph.OpenDB(dbPath)
	if err != nil {
		slog.Error("failed to open database", "error", err, "path", dbPath)
		os.Exit(1)
	}
	defer db.Close()

	// Load decay config
	decayCfg, err := config.LoadDecayConfig(ctx, db)
	if err != nil {
		slog.Warn("failed to load decay config, using defaults", "error", err)
		decayCfg = config.DefaultDecayConfig()
	}

	store := graph.NewStore(db)
	store.HalfLifeDays = decayCfg.HalfLifeDays
	store.ReinforceBump = decayCfg.ReinforcementBump
	store.StrengthCap = decayCfg.StrengthCap

	// Leader election — only the leader runs dashboard + housekeeping
	lockPath, err := config.LockPath()
	if err != nil {
		slog.Warn("failed to resolve lock path, skipping leader election", "error", err)
	} else {
		leader := singleton.New(lockPath)
		isLeader, err := leader.TryAcquire(os.Getpid(), 0)
		if err != nil {
			slog.Warn("leader election failed", "error", err)
		} else if isLeader {
			defer leader.Release()
			go store.RunHousekeeping(ctx, 1*time.Hour)

			// Start dashboard
			dash := dashboard.New(store)
			port, err := dash.Start(0)
			if err != nil {
				slog.Warn("failed to start dashboard", "error", err)
			} else {
				// Re-acquire with the dashboard port so followers can find it
				leader.TryAcquire(os.Getpid(), port)
			}
		}
	}

	slog.Info("starting lethe MCP server",
		"version", version,
		"transport", "stdio",
		"db", dbPath,
		"half_life_days", decayCfg.HalfLifeDays,
	)

	s := mcpserver.New(version, mcpserver.WithStore(store))
	if err := s.Run(ctx); err != nil {
		slog.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}

func openDashboard() {
	// Try to read the leader's port from the lockfile
	lockPath, err := config.LockPath()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: could not resolve lock path:", err)
		os.Exit(1)
	}

	leader := singleton.New(lockPath)
	_, port, err := leader.LeaderInfo()
	if err != nil || port == 0 {
		fmt.Fprintln(os.Stderr, "No running lethe process found with a dashboard. Start lethe with --stdio first.")
		os.Exit(1)
	}

	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	fmt.Fprintf(os.Stderr, "Opening dashboard at %s\n", url)
	openBrowser(url)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		fmt.Fprintln(os.Stderr, "Open this URL in your browser:", url)
		return
	}
	_ = cmd.Start()
}
