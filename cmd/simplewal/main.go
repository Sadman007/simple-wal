package main

import (
	"fmt"
	"time"

	"github.com/Sadman007/simplewal/internal/wal"
)

func main() {
	wal, _ := wal.InitWAL(wal.CreateDefaultWALConfig("/home/sadmansakib/code/simplewal"))
	if err := wal.WriteEntryWithCheckpoint([]byte("test 11")); err != nil {
		fmt.Printf("failed to write entry: %v\n", err)
	}

	if err := wal.WriteEntry([]byte("42")); err != nil {
		fmt.Printf("failed to write entry: %v\n", err)
	}

	time.Sleep(301 * time.Millisecond)

	entries, _ := wal.ReadAll()
	fmt.Printf("Read %d entries from the WAL:\n", len(entries))
	for _, entry := range entries {
		fmt.Printf("LogSeqNumber: %d, Data: %s, IsCheckpoint: %t\n", entry.LogSeqNumber, entry.Data, entry.IsCheckpoint)
	}
}
