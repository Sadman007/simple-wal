package wal

import (
	"bufio"
	"context"
	"os"
	"sync"
	"time"
)

const (
	SyncInterval  = 300 * time.Millisecond // Interval to fsync data to disk
	SegmentPrefix = "wal-segment-"         // Prefix for WAL segment filenames
)

// WAL represents the write-ahead log structure.
type WAL struct {
	// File and directory management
	directory      string        // Directory where WAL segment files are stored
	currSegment    *os.File      // Current segment file being written
	bufWriter      *bufio.Writer // Buffered writer for currSegment
	currSegmentIdx int           // Index of the current segment

	// Sequence tracking
	lastSeqNo uint64 // Last written sequence number

	// File size and rotation control
	maxFileSize uint64 // Max size in bytes for a single segment file
	maxSegments int    // Max number of WAL segments to retain

	// Synchronization
	lock        sync.Mutex  // Mutex to protect concurrent access to WAL
	syncTimer   *time.Timer // Timer to trigger periodic syncing
	shouldFsync bool        // Whether fsync is enabled for durability

	// Lifecycle control
	ctx    context.Context    // Context for cancellation and lifecycle management
	cancel context.CancelFunc // Function to cancel the WAL's background tasks
}
