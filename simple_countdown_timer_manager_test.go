package simplecountdowntimermanager

import (
	"errors"
	"testing"
	"time"
)

// fakeClock is a deterministic clock for tests.
type fakeClock struct {
	now time.Time
}

func (f *fakeClock) Now() time.Time { return f.now }
func (f *fakeClock) Advance(d time.Duration) {
	f.now = f.now.Add(d)
}

func newTestManager() (*Manager, *fakeClock) {
	fc := &fakeClock{now: time.Unix(0, 0)}
	return New(fc.Now), fc
}

func TestRemainingBeforeStartReturnsFullDuration(t *testing.T) {
	m, _ := newTestManager()
	m.Add("a", 5*time.Second)

	got, err := m.Remaining("a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 5*time.Second {
		t.Fatalf("got %v, want %v", got, 5*time.Second)
	}
}

func TestRemainingAfterStartCountsDown(t *testing.T) {
	m, fc := newTestManager()
	m.Add("a", 10*time.Second)
	if err := m.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	fc.Advance(3 * time.Second)
	got, err := m.Remaining("a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 7*time.Second {
		t.Fatalf("got %v, want %v", got, 7*time.Second)
	}
}

func TestRemainingClampsAtZero(t *testing.T) {
	m, fc := newTestManager()
	m.Add("a", 2*time.Second)
	if err := m.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	fc.Advance(5 * time.Second)
	got, err := m.Remaining("a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 0 {
		t.Fatalf("got %v, want 0", got)
	}
}

func TestRemainingUnknownTimer(t *testing.T) {
	m, _ := newTestManager()
	_, err := m.Remaining("nope")
	if !errors.Is(err, ErrUnknownTimer) {
		t.Fatalf("got %v, want ErrUnknownTimer", err)
	}
}

func TestStartTwiceReturnsError(t *testing.T) {
	m, _ := newTestManager()
	m.Add("a", time.Second)
	if err := m.Start(); err != nil {
		t.Fatalf("first start: %v", err)
	}
	if err := m.Start(); !errors.Is(err, ErrAlreadyStarted) {
		t.Fatalf("second start: got %v, want ErrAlreadyStarted", err)
	}
}

func TestAddNegativeDurationClampsToZero(t *testing.T) {
	m, _ := newTestManager()
	m.Add("a", -5*time.Second)
	got, err := m.Remaining("a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 0 {
		t.Fatalf("got %v, want 0", got)
	}
}

func TestAddReplacesExistingDuration(t *testing.T) {
	m, fc := newTestManager()
	m.Add("a", 10*time.Second)
	m.Add("a", 4*time.Second)
	if err := m.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	fc.Advance(time.Second)
	got, err := m.Remaining("a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 3*time.Second {
		t.Fatalf("got %v, want %v", got, 3*time.Second)
	}
}

func TestExpiredBeforeStartIsFalse(t *testing.T) {
	m, _ := newTestManager()
	m.Add("a", 0)
	expired, err := m.Expired("a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if expired {
		t.Fatal("expected not expired before start")
	}
}

func TestExpiredAfterElapsed(t *testing.T) {
	m, fc := newTestManager()
	m.Add("a", 2*time.Second)
	if err := m.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	fc.Advance(2 * time.Second)
	expired, err := m.Expired("a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !expired {
		t.Fatal("expected expired")
	}
}

func TestExpiredUnknownTimer(t *testing.T) {
	m, _ := newTestManager()
	_, err := m.Expired("nope")
	if !errors.Is(err, ErrUnknownTimer) {
		t.Fatalf("got %v, want ErrUnknownTimer", err)
	}
}

func TestNamesReturnsCopy(t *testing.T) {
	m, _ := newTestManager()
	m.Add("a", time.Second)
	m.Add("b", time.Second)

	names := m.Names()
	if len(names) != 2 {
		t.Fatalf("got %d names, want 2", len(names))
	}

	// Mutating the returned slice must not affect the manager.
	for i := range names {
		names[i] = "mutated"
	}
	again := m.Names()
	for _, n := range again {
		if n == "mutated" {
			t.Fatal("Names returned a slice backed by the manager")
		}
	}
}

func TestClockMovingBackwardsClampsToFullDuration(t *testing.T) {
	m, fc := newTestManager()
	m.Add("a", 5*time.Second)
	if err := m.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	// Move the clock backwards; remaining should not exceed the original.
	fc.now = fc.now.Add(-time.Second)
	got, err := m.Remaining("a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 5*time.Second {
		t.Fatalf("got %v, want %v", got, 5*time.Second)
	}
}
