// Package wal provides a simple disk-backed Write-Ahead Log for the agent.
// Metrics are appended to the WAL when the network is unavailable and replayed
// in order when connectivity is restored. This prevents data loss during
// temporary outages.
package wal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	maxEntries = 2000 // drop oldest beyond this limit
	walFile    = "agent-wal.jsonl"
)

type WAL struct {
	mu   sync.Mutex
	path string
}

func New(dir string) (*WAL, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("wal mkdir: %w", err)
	}
	return &WAL{path: filepath.Join(dir, walFile)}, nil
}

// Append serialises v and appends it as a JSON line to the WAL.
func (w *WAL) Append(v interface{}) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(w.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "%s\n", data)
	return err
}

// Drain returns all buffered entries, decoded into values of type T,
// then truncates the WAL.
func Drain[T any](w *WAL) ([]T, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	f, err := os.Open(w.path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []T
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var v T
		if err := json.Unmarshal(scanner.Bytes(), &v); err != nil {
			continue // skip malformed lines
		}
		entries = append(entries, v)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	f.Close()

	// Keep only the newest entries if we overflowed
	if len(entries) > maxEntries {
		entries = entries[len(entries)-maxEntries:]
	}

	// Truncate WAL after drain
	_ = os.Truncate(w.path, 0)
	return entries, nil
}

// Size returns the number of pending entries.
func (w *WAL) Size() int {
	w.mu.Lock()
	defer w.mu.Unlock()

	f, err := os.Open(w.path)
	if err != nil {
		return 0
	}
	defer f.Close()

	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		count++
	}
	return count
}
