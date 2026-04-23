package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
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

const defaultDashboardPort = 52117 // stable port so URL doesn't change across restarts

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
		runDashboard()
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

	dbPath, store := openStore(ctx)

	// Leader election — only the leader runs dashboard + housekeeping
	lockPath, err := config.LockPath()
	if err != nil {
		slog.Warn("failed to resolve lock path, skipping leader election", "error", err)
	} else {
		leader := singleton.New(lockPath)
		isLeader, err := leader.TryAcquire(os.Getpid())
		if err != nil {
			slog.Warn("leader election failed", "error", err)
		} else if isLeader {
			defer leader.Release()
			go store.RunHousekeeping(ctx, 1*time.Hour)

			// Start dashboard — try stable port first, fall back to random
			dash := dashboard.New(store)
			port, err := dash.Start(defaultDashboardPort)
			if err != nil {
				port, err = dash.Start(0)
			}
			if err != nil {
				slog.Warn("failed to start dashboard", "error", err)
			} else {
				leader.SetPort(os.Getpid(), port)
			}
		}
	}

	slog.Info("starting lethe MCP server",
		"version", version,
		"transport", "stdio",
		"db", dbPath,
	)

	s := mcpserver.New(version, mcpserver.WithStore(store))
	if err := s.Run(ctx); err != nil {
		slog.Error("server exited with error", "error", err)
		os.Exit(1)
	}
	// DB stays open via the store for the lifetime of the process
}

// runDashboard either connects to an existing leader's dashboard,
// or starts its own standalone dashboard server if no leader is running.
func runDashboard() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	lockPath, err := config.LockPath()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: could not resolve lock path:", err)
		os.Exit(1)
	}

	// Check if a leader is already running with a dashboard
	leader := singleton.New(lockPath)
	_, port, err := leader.LeaderInfo()
	if err == nil && port > 0 {
		// Verify it's actually reachable
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/api/stats", port))
		if err == nil {
			resp.Body.Close()
			url := fmt.Sprintf("http://127.0.0.1:%d", port)
			fmt.Fprintf(os.Stderr, "Dashboard already running at %s\n", url)
			openBrowser(url)
			return
		}
	}

	// No live dashboard found — start our own standalone server
	fmt.Fprintln(os.Stderr, "No running dashboard found. Starting standalone dashboard...")

	_, store := openStore(ctx)

	// Try to become leader (for housekeeping)
	isLeader, _ := leader.TryAcquire(os.Getpid())
	if isLeader {
		defer leader.Release()
		go store.RunHousekeeping(ctx, 1*time.Hour)
	}

	dash := dashboard.New(store)
	port, err = dash.Start(defaultDashboardPort)
	if err != nil {
		port, err = dash.Start(0)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: failed to start dashboard:", err)
		os.Exit(1)
	}

	if isLeader {
		leader.SetPort(os.Getpid(), port)
	}

	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	fmt.Fprintf(os.Stderr, "Dashboard running at %s\n", url)
	fmt.Fprintln(os.Stderr, "Press Ctrl+C to stop.")
	openBrowser(url)

	// Block until interrupted
	<-ctx.Done()
	dash.Stop(context.Background())
}

// openStore opens the database and returns a configured store.
func openStore(ctx context.Context) (string, *graph.Store) {
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

	decayCfg, err := config.LoadDecayConfig(ctx, db)
	if err != nil {
		slog.Warn("failed to load decay config, using defaults", "error", err)
		decayCfg = config.DefaultDecayConfig()
	}

	store := graph.NewStore(db)
	store.HalfLifeDays = decayCfg.HalfLifeDays
	store.ReinforceBump = decayCfg.ReinforcementBump
	store.StrengthCap = decayCfg.StrengthCap
	return dbPath, store
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
