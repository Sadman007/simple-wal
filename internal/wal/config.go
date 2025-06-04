package wal

import "time"

type WALConfig struct {
	Directory    string
	EnableFsync  bool
	MaxFileSize  uint64
	MaxSegments  int
	SyncInterval time.Duration
}

func CreateDefaultWALConfig(directory string) WALConfig {
	return WALConfig{
		Directory:    directory,
		EnableFsync:  true,
		MaxFileSize:  16 * 1024 * 1024,
		MaxSegments:  1, // Currently we are not rotating segments. Replace with a higher number later.
		SyncInterval: 300 * time.Millisecond,
	}
}
