package wal

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Sadman007/simplewal/internal/wal"
)

func TestInitWAL_CreatesSegmentFile(t *testing.T) {
	dir := t.TempDir()
	cfg := wal.CreateDefaultWALConfig(dir)
	w, err := wal.InitWAL(cfg)
	if err != nil {
		t.Fatalf("InitWAL failed: %v", err)
	}
	defer w.Close()

	segmentPath := filepath.Join(dir, wal.SegmentPrefix+"0")
	if _, err := os.Stat(segmentPath); os.IsNotExist(err) {
		t.Errorf("segment file was not created")
	}
}

func TestWAL_WriteAndReadEntry(t *testing.T) {
	dir := t.TempDir()
	cfg := wal.CreateDefaultWALConfig(dir)
	w, err := wal.InitWAL(cfg)
	if err != nil {
		t.Fatalf("InitWAL failed: %v", err)
	}
	defer w.Close()

	for i := 1; i <= 3; i++ {
		data := []byte(fmt.Sprintf("hello wal %d", i))
		if err := w.WriteEntry(data); err != nil {
			t.Fatalf("WriteEntry failed: %v", err)
		}
	}
	w.Sync()

	entries, err := w.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("ReadAll returned wrong number of entries. Expected %d, Got %d", 3, len(entries))
	}
	for i := 1; i <= 3; i++ {
		expectedData := fmt.Sprintf("hello wal %d", i)
		if string(entries[i-1].Data) != expectedData {
			t.Errorf("ReadAll entry %d data mismatch. Expected %s, Got %s", i, expectedData, entries[i-1].Data)
		}
	}
}

func TestWAL_WriteWithCheckpointAndReadEntry(t *testing.T) {
	dir := t.TempDir()
	cfg := wal.CreateDefaultWALConfig(dir)
	w, err := wal.InitWAL(cfg)
	if err != nil {
		t.Fatalf("InitWAL failed: %v", err)
	}
	defer w.Close()

	for i := 1; i <= 3; i++ {
		data := []byte(fmt.Sprintf("hello wal %d", i))
		if err := w.WriteEntryWithCheckpoint(data); err != nil {
			t.Fatalf("WriteEntry failed: %v", err)
		}
	}
	w.Sync()

	entries, err := w.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("ReadAll returned wrong number of entries. Expected %d, Got %d", 3, len(entries))
	}
	for i := 1; i <= 3; i++ {
		expectedData := fmt.Sprintf("hello wal %d", i)
		if string(entries[i-1].Data) != expectedData && !entries[i-1].GetIsCheckpoint() {
			t.Errorf("ReadAll entry %d data mismatch. Expected %s, Got %s", i, expectedData, entries[i-1].Data)
		}
	}
}

func TestWAL_CRCVerification(t *testing.T) {
	dir := t.TempDir()
	cfg := wal.CreateDefaultWALConfig(dir)
	w, err := wal.InitWAL(cfg)
	if err != nil {
		t.Fatalf("InitWAL failed: %v", err)
	}
	defer w.Close()

	data := []byte("crc test")
	if err := w.WriteEntry(data); err != nil {
		t.Fatalf("WriteEntry failed: %v", err)
	}
	w.Sync()

	// Corrupt the file
	segmentPath := filepath.Join(dir, wal.SegmentPrefix+"0")
	f, err := os.OpenFile(segmentPath, os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("failed to open segment: %v", err)
	}
	defer f.Close()
	if _, err := f.WriteAt([]byte{0xFF}, 8); err != nil {
		t.Fatalf("failed to corrupt segment: %v", err)
	}

	_, err = w.ReadAll()
	if err == nil || err.Error() == "" {
		t.Errorf("expected CRC error, got nil")
	}
}
