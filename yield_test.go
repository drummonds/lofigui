package lofigui

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestYieldDoesNotPanic(t *testing.T) {
	// Basic smoke test: Yield must not panic on any platform.
	Yield()
}

func TestYieldAllowsConcurrentProgress(t *testing.T) {
	// Simulates the gobank scenario: a tight loop in one goroutine
	// should not starve another goroutine that needs to run.
	var counter atomic.Int64
	done := make(chan struct{})

	// "Background model" goroutine — tight loop with Yield.
	go func() {
		for i := 0; i < 1000; i++ {
			counter.Add(1)
			Yield()
		}
		close(done)
	}()

	// Give the loop a moment to start, then check progress from
	// the "UI" goroutine.  Without Yield (or with a no-op Gosched
	// in WASM) the counter might not advance until the loop finishes.
	time.Sleep(5 * time.Millisecond)

	select {
	case <-done:
		// Loop already finished — that's fine on native.
	default:
		// Still running: verify some progress was made (i.e. the
		// tight loop yielded and let us observe an intermediate value).
		if counter.Load() == 0 {
			t.Error("background loop made no progress; Yield may not be yielding")
		}
	}

	// Wait for completion with a generous timeout.
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("background loop did not complete within timeout")
	}

	if counter.Load() != 1000 {
		t.Errorf("expected counter=1000, got %d", counter.Load())
	}
}

func TestYieldInBufferLoop(t *testing.T) {
	// Realistic usage: a model function that writes to the lofigui
	// buffer in a loop, calling Yield each iteration so WASM builds
	// can repaint between writes.
	Reset()

	for i := 0; i < 50; i++ {
		Printf("Item %d", i)
		Yield()
	}

	buf := Buffer()
	if buf == "" {
		t.Fatal("buffer is empty after loop with Yield")
	}

	// Verify first and last items are present.
	if !contains(buf, "Item 0") {
		t.Error("buffer missing 'Item 0'")
	}
	if !contains(buf, "Item 49") {
		t.Error("buffer missing 'Item 49'")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
