package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/MrMoneyInTheBank/jobit/internal/store"
)

// The OS releases the advisory lock even after a crash. Keep the lock file:
// unlinking it would allow separate processes to lock different inodes.
func acquireInstance() (*os.File, error) {
	databasePath, err := store.DefaultDatabasePath()
	if err != nil {
		return nil, err
	}
	directory := filepath.Dir(databasePath)
	if err = os.MkdirAll(directory, 0o700); err != nil {
		return nil, err
	}
	lockFile, err := os.OpenFile(filepath.Join(directory, "jobit.lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		lockFile.Close()
		if err == syscall.EWOULDBLOCK || err == syscall.EAGAIN {
			return nil, fmt.Errorf("Jobit is already running for this user. Return to its terminal tab/window, or quit that session before starting another. Attaching to another terminal session is not supported")
		}
		return nil, fmt.Errorf("lock Jobit instance: %w", err)
	}
	return lockFile, nil
}
