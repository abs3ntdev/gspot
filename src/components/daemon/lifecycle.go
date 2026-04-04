package daemon

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/abs3ntdev/gspot/src/config"
)

// IsRunning checks if a daemon process is alive by reading the PID file
// and sending signal 0. Returns the PID if running, 0 otherwise.
func IsRunning(conf *config.Config) (int, bool) {
	data, err := os.ReadFile(conf.PidFile)
	if err != nil {
		return 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return 0, false
	}
	// Signal 0 checks if process exists without actually signaling it.
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		// Process doesn't exist, clean up stale files.
		os.Remove(conf.PidFile)
		return 0, false
	}
	return pid, true
}

// WritePid writes the current process PID to the PID file.
func WritePid(conf *config.Config) error {
	return os.WriteFile(conf.PidFile, []byte(strconv.Itoa(os.Getpid())), 0o644)
}

// RemovePid removes the PID file.
func RemovePid(conf *config.Config) {
	os.Remove(conf.PidFile)
}

// Start launches the daemon as a detached background process by re-execing
// the current binary with "daemon run". Returns the child PID.
func Start(conf *config.Config) (int, error) {
	if pid, running := IsRunning(conf); running {
		return pid, fmt.Errorf("daemon already running (pid %d)", pid)
	}

	exe, err := os.Executable()
	if err != nil {
		return 0, fmt.Errorf("could not determine executable path: %w", err)
	}

	cmd := exec.Command(exe, "daemon", "run")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // Detach from parent session.
	}
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start daemon: %w", err)
	}

	// Release the child so it isn't waited on by this process.
	pid := cmd.Process.Pid
	cmd.Process.Release()

	// Wait for the daemon to become ready by polling the socket.
	if err := waitForSocket(conf.SocketPath, 5*time.Second); err != nil {
		return pid, fmt.Errorf("daemon started (pid %d) but socket not ready: %w", pid, err)
	}

	return pid, nil
}

// Stop sends SIGTERM to the running daemon.
func Stop(conf *config.Config) error {
	pid, running := IsRunning(conf)
	if !running {
		return fmt.Errorf("daemon is not running")
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("could not find process %d: %w", pid, err)
	}
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to stop daemon (pid %d): %w", pid, err)
	}
	// Wait briefly for cleanup.
	for range 20 {
		time.Sleep(100 * time.Millisecond)
		if _, alive := IsRunning(conf); !alive {
			return nil
		}
	}
	return nil
}

// EnsureRunning checks if the daemon is running. If not, it starts one
// and waits for it to be ready. This is called automatically by RPC clients.
func EnsureRunning(conf *config.Config) error {
	if _, running := IsRunning(conf); running {
		return nil
	}
	_, err := Start(conf)
	return err
}

// waitForSocket polls until the Unix socket is accepting connections.
func waitForSocket(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("unix", path, 200*time.Millisecond)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for socket %s", path)
}
