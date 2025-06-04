package wal

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	pb "github.com/Sadman007/simplewal/proto"
	"google.golang.org/protobuf/proto"
)

// prepareCurrentSegment ensures that a log segment file exists in the specified directory.
// It scans the directory for files matching the segment prefix, determines the highest segment ID present,
// and returns it. If no segment files are found, it creates a new segment file with ID 0 and returns 0.
// Returns the maximum segment ID found (or 0 if a new segment is created), or an error if the operation fails.
func prepareCurrentSegment(dir string) (int, error) {
	// Get the list of log segment files in the dir.
	files, err := filepath.Glob(filepath.Join(dir, SegmentPrefix+"*"))
	if err != nil {
		return -1, err
	}

	// If no segment is present, create one.
	if len(files) == 0 {
		file, err := createSegmentFile(dir, 0)
		if err != nil {
			return -1, err
		}
		if err := file.Close(); err != nil {
			return -1, err
		}
		return 0, nil
	}

	var maxSegmentID int
	for _, file := range files {
		filename := filepath.Base(file)

		nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))

		if !strings.HasPrefix(nameWithoutExt, SegmentPrefix) {
			continue
		}

		idStr := strings.TrimPrefix(nameWithoutExt, SegmentPrefix)
		if idStr == "" {
			continue
		}

		segmentID, err := strconv.Atoi(idStr)
		if err != nil {
			return -1, fmt.Errorf(
				"file %q matched prefix %q, but failed to parse segment ID from %q (derived from %q): %w",
				file, SegmentPrefix, idStr, nameWithoutExt, err,
			)
		}
		maxSegmentID = max(maxSegmentID, segmentID)
	}

	return maxSegmentID, nil
}

// createSegmentFile creates a new segment file in the specified directory with the given segment ID.
func createSegmentFile(directory string, segmentID int) (*os.File, error) {
	fileName := fmt.Sprintf("%s%d", SegmentPrefix, segmentID)
	filePath := filepath.Join(directory, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create segment file at %s: %w", filePath, err)
	}

	return file, nil
}

// getLastSeqNo returns the last sequence number in the current log.
func (wal *WAL) getLastSeqNo() (uint64, error) {
	entry, err := wal.getLastEntryInLog()
	if err != nil {
		return 0, err
	}

	// If no entry is found and there's no error, we simply return 0 as the last sequence number.
	if entry == nil {
		return 0, nil
	}

	return entry.GetLogSeqNumber(), nil
}

// getLastEntryInLog iterates through the current segment file and returns the last valid entry.
func (wal *WAL) getLastEntryInLog() (*pb.WAL_Entry, error) {
	file, err := os.OpenFile(wal.currSegment.Name(), os.O_RDONLY, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			// If the file doesn't exist, there are no entries.
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var lastEntry *pb.WAL_Entry

	for {
		// Read the size of the next entry.
		var size int32
		if err := binary.Read(file, binary.LittleEndian, &size); err != nil {
			if err == io.EOF {
				// Clean end-of-file. We've read all entries.
				// The last valid entry is what we have in lastEntry.
				break
			}
			// Any other error (e.g., io.ErrUnexpectedEOF) means the file is
			// likely truncated or corrupt. We should stop and return the last
			// known good entry.
			return lastEntry, nil
		}

		// According to the size, read the entry data.
		// Using io.ReadFull ensures that we get an error if the file ends
		// prematurely (truncated entry).
		data := make([]byte, size)
		if _, err := io.ReadFull(file, data); err != nil {
			return lastEntry, nil
		}

		entry, err := unmarshalAndVerifyEntry(data)
		if err != nil {
			// The entry is corrupt (e.g., CRC mismatch).
			// Return the last valid entry we successfully read before this one.
			return lastEntry, nil
		}

		lastEntry = entry
	}

	return lastEntry, nil
}

// unmarshalAndVerifyEntry unmarshals a WAL entry from the provided byte slice
// and verifies its CRC. If successful, it returns the entry; otherwise, it returns an error.
func unmarshalAndVerifyEntry(data []byte) (*pb.WAL_Entry, error) {
	var entry pb.WAL_Entry
	if err := proto.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal entry: %w", err)
	}

	if !verifyCRC(&entry) {
		return nil, fmt.Errorf("crc verification failed for entry LSN %d", entry.GetLogSeqNumber())
	}
	return &entry, nil
}

// verifyCRC checks if the CRC of the entry matches the computed CRC from its data.
func verifyCRC(entry *pb.WAL_Entry) bool {
	return entry.GetCRC() == crc32.ChecksumIEEE(entry.GetData())
}
