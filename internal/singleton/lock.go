package singleton

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/gofrs/flock"
)

// Leader manages singleton leader election via a lockfile.
// The leader runs the dashboard and housekeeping; followers skip them.
type Leader struct {
	flock    *flock.Flock
	lockPath string
	isLeader bool
	mu       sync.Mutex
}

// New creates a new Leader for the given lockfile path.
func New(lockPath string) *Leader {
	return &Leader{
		flock:    flock.New(lockPath),
		lockPath: lockPath,
	}
}

// TryAcquire attempts to become the leader via a non-blocking lock.
// Returns true if this process is now the leader.
func (l *Leader) TryAcquire(pid int) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	locked, err := l.flock.TryLock()
	if err != nil {
		return false, fmt.Errorf("lock: %w", err)
	}
	l.isLeader = locked

	if locked {
		l.writeInfo(pid, 0)
		slog.Info("acquired leader lock", "pid", pid)
	} else {
		slog.Info("another process is leader, running as follower")
	}

	return locked, nil
}

// SetPort updates the info file with the dashboard port.
// Call this after the dashboard has started and you know the port.
func (l *Leader) SetPort(pid, port int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.writeInfo(pid, port)
}

func (l *Leader) writeInfo(pid, port int) {
	content := fmt.Sprintf("%d\n%d\n", pid, port)
	if err := os.WriteFile(l.lockPath+".info", []byte(content), 0o644); err != nil {
		slog.Warn("failed to write lock info", "error", err)
	}
}

// IsLeader returns whether this process currently holds the lock.
func (l *Leader) IsLeader() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.isLeader
}

// LeaderInfo reads the PID and port of the current leader from the info file.
func (l *Leader) LeaderInfo() (pid int, port int, err error) {
	data, err := os.ReadFile(l.lockPath + ".info")
	if err != nil {
		return 0, 0, err
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 2 {
		return 0, 0, fmt.Errorf("malformed lock info")
	}
	pid, err = strconv.Atoi(lines[0])
	if err != nil {
		return 0, 0, fmt.Errorf("parse pid: %w", err)
	}
	port, err = strconv.Atoi(lines[1])
	if err != nil {
		return 0, 0, fmt.Errorf("parse port: %w", err)
	}
	return pid, port, nil
}

// Release releases the leader lock and cleans up the info file.
func (l *Leader) Release() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.isLeader = false
	_ = os.Remove(l.lockPath + ".info")
	return l.flock.Unlock()
}
