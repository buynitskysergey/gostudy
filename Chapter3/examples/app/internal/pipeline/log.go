package pipeline

import (
	"fmt"
	"sync"
	"time"
)

var (
	programStart = time.Now()
	logMu        sync.Mutex
)

// Log выводит сообщение с меткой времени (ms от старта) и источником.
func Log(source, message string) {
	logMu.Lock()
	defer logMu.Unlock()
	elapsed := time.Since(programStart).Milliseconds()
	fmt.Printf("%dms | [%s]: %s\n", elapsed, source, message)
}
