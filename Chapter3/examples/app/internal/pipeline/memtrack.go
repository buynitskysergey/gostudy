package pipeline

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// FormatBytes форматирует размер в человекочитаемый вид.
func FormatBytes(b uint64) string {
	switch {
	case b >= 1<<20:
		return fmt.Sprintf("%.2f MiB", float64(b)/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.2f KiB", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

// StartPeakMemTracker запускает goroutine, которая периодически снимает runtime.MemStats
// и запоминает максимум HeapInuse (куча) и Sys (всего от OS, включая стеки goroutines).
// Возвращает функцию stop: вызови её в конце main, чтобы получить зафиксированный пик.
func StartPeakMemTracker() func() (heapPeak, sysPeak uint64) {
	stop := make(chan struct{})
	var wg sync.WaitGroup
	var mu sync.Mutex
	var peakHeap, peakSys uint64

	sample := func(ms *runtime.MemStats) {
		runtime.ReadMemStats(ms)
		mu.Lock()
		if ms.HeapInuse > peakHeap {
			peakHeap = ms.HeapInuse
		}
		if ms.Sys > peakSys {
			peakSys = ms.Sys
		}
		mu.Unlock()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		var ms runtime.MemStats
		ticker := time.NewTicker(1 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				sample(&ms)
			case <-stop:
				sample(&ms)
				return
			}
		}
	}()

	return func() (uint64, uint64) {
		close(stop)
		wg.Wait()
		mu.Lock()
		defer mu.Unlock()
		return peakHeap, peakSys
	}
}
