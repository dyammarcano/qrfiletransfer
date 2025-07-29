package testutils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"testing"
	"time"
)

type TestProfile struct {
	TestName         string        `json:"test_name"`
	Timestamp        string        `json:"timestamp"`
	Duration         time.Duration `json:"duration"`
	AllocDelta       int64         `json:"alloc_delta"`
	TotalAllocDelta  int64         `json:"total_alloc_delta"`
	AllocsDelta      int64         `json:"allocs_delta"`
	FreesDelta       int64         `json:"frees_delta"`
	HeapAllocDelta   int64         `json:"heap_alloc_delta"`
	HeapObjectsDelta int64         `json:"heap_objects_delta"`
	GCCycles         int64         `json:"gc_cycles"`
	GCCPUFraction    float64       `json:"gc_cpu_fraction"`
	StackInuseDelta  int64         `json:"stack_inuse_delta"`
	StackSysDelta    int64         `json:"stack_sys_delta"`
}

func MeasureTestPerformance(t *testing.T) func() {
	t.Helper()

	var mStart runtime.MemStats
	runtime.ReadMemStats(&mStart)

	start := time.Now()
	testName := t.Name()

	return func() {
		var mEnd runtime.MemStats
		runtime.ReadMemStats(&mEnd)

		duration := time.Since(start)

		profile := TestProfile{
			TestName:         testName,
			Timestamp:        time.Now().Format(time.RFC3339),
			Duration:         duration,
			AllocDelta:       int64(mEnd.Alloc) - int64(mStart.Alloc),
			TotalAllocDelta:  int64(mEnd.TotalAlloc) - int64(mStart.TotalAlloc),
			AllocsDelta:      int64(mEnd.Mallocs) - int64(mStart.Mallocs),
			FreesDelta:       int64(mEnd.Frees) - int64(mStart.Frees),
			HeapAllocDelta:   int64(mEnd.HeapAlloc) - int64(mStart.HeapAlloc),
			HeapObjectsDelta: int64(mEnd.HeapObjects) - int64(mStart.HeapObjects),
			GCCycles:         int64(mEnd.NumGC) - int64(mStart.NumGC),
			GCCPUFraction:    mEnd.GCCPUFraction,
			StackInuseDelta:  int64(mEnd.StackInuse) - int64(mStart.StackInuse),
			StackSysDelta:    int64(mEnd.StackSys) - int64(mStart.StackSys),
		}

		// Pretty print to test logs
		t.Logf(`
Test Duration: %s
Memory Allocations:
  Current memory in use: %d bytes
  Total memory allocated: %d bytes
  Allocation operations: %d allocs, %d frees (net: %d)
Heap Statistics:
  Heap allocation delta: %d bytes
  Heap objects delta: %d objects
Garbage Collection:
  GC cycles: %d
  GC CPU fraction: %.2f%%
Stack Statistics:
  Stack in use delta: %d bytes
  Stack system delta: %d bytes`,
			profile.Duration,
			profile.AllocDelta,
			profile.TotalAllocDelta,
			profile.AllocsDelta,
			profile.FreesDelta,
			profile.AllocsDelta-profile.FreesDelta,
			profile.HeapAllocDelta,
			profile.HeapObjectsDelta,
			profile.GCCycles,
			profile.GCCPUFraction*100,
			profile.StackInuseDelta,
			profile.StackSysDelta,
		)

		outputDir := filepath.Join("../profiles", time.Now().Format("20060102_150405"))
		_ = os.MkdirAll(outputDir, 0755)

		file, err := os.Create(filepath.Join(outputDir, fmt.Sprintf("%s.json", testName)))
		if err != nil {
			t.Logf("Failed to save profile: %v", err)
			return
		}
		defer func(file *os.File) {
			if err := file.Close(); err != nil {
				return
			}
		}(file)

		_ = json.NewEncoder(file).Encode(profile)

		// Guardar perfiles pprof estándar
		heapFile, _ := os.Create(filepath.Join(outputDir, fmt.Sprintf("%s_heap.prof", testName)))
		_ = pprof.WriteHeapProfile(heapFile)
		if err := heapFile.Close(); err != nil {
			return
		}

		// Profile de goroutines (útil para ver leaks o bloqueos)
		goroutineFile, _ := os.Create(filepath.Join(outputDir, fmt.Sprintf("%s_goroutines.prof", testName)))
		if err := pprof.Lookup("goroutine").WriteTo(goroutineFile, 0); err != nil {
			return
		}
		if err := goroutineFile.Close(); err != nil {
			return
		}

		// Profile de asignaciones (más útil con `pprof -alloc_space`)
		allocsFile, _ := os.Create(filepath.Join(outputDir, fmt.Sprintf("%s_allocs.prof", testName)))
		if err := pprof.Lookup("allocs").WriteTo(allocsFile, 0); err != nil {
			return
		}
		if err := allocsFile.Close(); err != nil {
			return
		}

		// CPU profile
		fCPU, err := os.Create(filepath.Join(outputDir, "cpu.prof"))
		if err != nil {
			t.Fatalf("could not create CPU profile: %v", err)
		}

		if err := pprof.StartCPUProfile(fCPU); err != nil {
			return
		}
		defer func() {
			pprof.StopCPUProfile()
			if err := fCPU.Close(); err != nil {
				return
			}
		}()

		// Memory profile (heap)
		defer func() {
			fMem, err := os.Create(filepath.Join(outputDir, "mem.prof"))
			if err != nil {
				t.Fatalf("could not create memory profile: %v", err)
			}
			defer func(fMem *os.File) {
				if err := fMem.Close(); err != nil {
					t.Fatalf("could not close memory profile: %v", err)
				}
			}(fMem)
			_ = pprof.WriteHeapProfile(fMem)
		}()
	}
}
