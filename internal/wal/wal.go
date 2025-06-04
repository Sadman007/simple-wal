package wal

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	pb "github.com/Sadman007/simplewal/proto"
	"google.golang.org/protobuf/proto"
)

func InitWAL(cfg WALConfig) (*WAL, error) {
	if cfg.Directory == "" {
		return nil, fmt.Errorf("WALConfig.Directory must not be empty")
	}

	if err := os.MkdirAll(cfg.Directory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create WAL directory: %w", err)
	}

	segmentId, err := prepareCurrentSegment(cfg.Directory)
	if err != nil {
		return nil, err
	}

	// Open the last segment file
	filePath := filepath.Join(cfg.Directory, fmt.Sprintf("%s%d", SegmentPrefix, segmentId))
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	// Seek to the end for appending
	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		return nil, fmt.Errorf("failed to seek to the end of segment: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	wal := &WAL{
		directory:      cfg.Directory,
		currSegment:    file,
		bufWriter:      bufio.NewWriter(file),
		lastSeqNo:      0,
		syncTimer:      time.NewTimer(cfg.SyncInterval),
		shouldFsync:    cfg.EnableFsync,
		maxFileSize:    cfg.MaxFileSize,
		maxSegments:    cfg.MaxSegments,
		currSegmentIdx: segmentId,
		ctx:            ctx,
		cancel:         cancel,
	}

	if wal.lastSeqNo, err = wal.getLastSeqNo(); err != nil {
		return nil, fmt.Errorf("failed to get last sequence number: %w", err)
	}

	go wal.keepSyncing()

	return wal, nil
}

// ReadAll reads all entries from the current WAL segment and returns them as a slice of WAL_Entry.
func (wal *WAL) ReadAll() ([]*pb.WAL_Entry, error) {
	wal.lock.Lock()
	defer wal.lock.Unlock()

	file, err := os.OpenFile(wal.currSegment.Name(), os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open WAL segment for reading: %w", err)
	}
	defer file.Close()

	var entries []*pb.WAL_Entry

	for {
		var size int32
		if err := binary.Read(file, binary.LittleEndian, &size); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to read entry size: %w", err)
		}

		data := make([]byte, size)
		if _, err := io.ReadFull(file, data); err != nil {
			return nil, fmt.Errorf("failed to read entry data: %w", err)
		}

		entry := &pb.WAL_Entry{}
		if err := proto.Unmarshal(data, entry); err != nil {
			return nil, fmt.Errorf("failed to unmarshal entry: %w", err)
		}

		if crc32.ChecksumIEEE(entry.GetData()) != entry.GetCRC() {
			return nil, fmt.Errorf("CRC mismatch for entry with seq no %d", entry.GetLogSeqNumber())
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// WriteEntry writes a new entry to the WAL with the provided data.
func (wal *WAL) WriteEntry(data []byte) error {
	wal.lock.Lock()
	defer wal.lock.Unlock()

	// TODO(Sadman007): Implement log segment rotation. Currently we are only using one segment.

	wal.lastSeqNo++
	entry := &pb.WAL_Entry{
		LogSeqNumber: wal.lastSeqNo,
		Data:         data,
		CRC:          crc32.ChecksumIEEE(data),
	}

	return wal.writeEntryToBuffer(entry)
}

// WriteEntryWithCheckpoint writes a new entry to the WAL with the provided data and marks it as a checkpoint.
func (wal *WAL) WriteEntryWithCheckpoint(data []byte) error {
	wal.lock.Lock()
	defer wal.lock.Unlock()

	// TODO(Sadman007): Implement log segment rotation. Currently we are only using one segment.

	if err := wal.Sync(); err != nil {
		return fmt.Errorf("failed to sync before writing checkpoint: %w", err)
	}

	wal.lastSeqNo++
	isCheckpoint := true
	entry := &pb.WAL_Entry{
		LogSeqNumber: wal.lastSeqNo,
		Data:         data,
		CRC:          crc32.ChecksumIEEE(data),
		IsCheckpoint: &isCheckpoint,
	}

	return wal.writeEntryToBuffer(entry)
}

// writeEntryToBuffer marshals the entry and writes it to the buffer.
func (wal *WAL) writeEntryToBuffer(entry *pb.WAL_Entry) error {
	marshaledEntry, err := proto.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal should never fail (%v)", err)
	}

	size := int32(len(marshaledEntry))

	// Write the size of the entry first
	if err := binary.Write(wal.bufWriter, binary.LittleEndian, size); err != nil {
		return fmt.Errorf("wal: failed to write entry size to buffer: %w", err)
	}

	// Then write the marshaled entry data
	if _, err := wal.bufWriter.Write(marshaledEntry); err != nil {
		return fmt.Errorf("wal: failed to write marshaled entry data to buffer: %w", err)
	}

	return nil
}

// keepSyncing runs in a separate goroutine to periodically sync the WAL to disk.
func (wal *WAL) keepSyncing() {
	for {
		select {
		case <-wal.ctx.Done():
			return
		case <-wal.syncTimer.C:
			wal.lock.Lock()
			err := wal.Sync()
			wal.lock.Unlock()
			if err != nil {
				log.Printf("failed to sync WAL: %v", err)
			}
		}
	}
}

// Sync flushes the buffer and fsyncs the current segment file to ensure data durability.
func (wal *WAL) Sync() error {
	if err := wal.bufWriter.Flush(); err != nil {
		return fmt.Errorf("failed to flush buffer: %w", err)
	}
	if wal.shouldFsync {
		if err := wal.currSegment.Sync(); err != nil {
			return fmt.Errorf("failed to fsync current segment: %w", err)
		}
	}

	wal.syncTimer.Reset(SyncInterval)

	return nil
}

// Close closes the WAL file. It also calls Sync() on the WAL.
func (wal *WAL) Close() error {
	wal.cancel()
	if err := wal.Sync(); err != nil {
		return err
	}
	return wal.currSegment.Close()
}
