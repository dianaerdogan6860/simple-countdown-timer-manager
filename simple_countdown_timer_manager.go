// Package simplecountdowntimermanager provides a small manager for named
// countdown timers that share a start time and report remaining time.
//
// The design is deliberately simple: a Manager holds a start instant and a
// clock function. Each timer is just a name plus a duration. Remaining time
// is computed on demand from the shared start time, so timers are always
// consistent with one another even if they are queried at different moments.
//
// The clock is injected so tests can be deterministic; production callers
// pass time.Now.
package simplecountdowntimermanager

import (
	"errors"
	"time"
)

// ErrUnknownTimer is returned when a timer name is not registered.
var ErrUnknownTimer = errors.New("unknown timer")

// ErrAlreadyStarted is returned by Start when the manager has already started.
var ErrAlreadyStarted = errors.New("manager already started")

// Manager tracks named countdown timers that share a start instant.
//
// Zero value is not usable; construct with New.
type Manager struct {
	clock    func() time.Time
	timers   map[string]time.Duration
	start    time.Time
	started  bool
}

// New returns a Manager that uses clock to read the current time.
// Pass time.Now in production and a fake clock in tests.
func New(clock func() time.Time) *Manager {
	if clock == nil {
		clock = time.Now
	}
	return &Manager{
		clock:  clock,
		timers: make(map[string]time.Duration),
	}
}

// Add registers a timer with the given name and duration.
// Adding the same name twice replaces the previous duration.
// Durations are clamped to zero: a negative duration is treated as zero.
func (m *Manager) Add(name string, d time.Duration) {
	if d < 0 {
		d = 0
	}
	m.timers[name] = d
}

// Start records the shared start instant. It returns ErrAlreadyStarted
// if Start has already been called successfully.
func (m *Manager) Start() error {
	if m.started {
		return ErrAlreadyStarted
	}
	m.start = m.clock()
	m.started = true
	return nil
}

// Started reports whether Start has been called successfully.
func (m *Manager) Started() bool {
	return m.started
}

// Names returns the registered timer names in an unspecified order.
// The slice is a fresh copy; mutating it does not affect the Manager.
func (m *Manager) Names() []string {
	out := make([]string, 0, len(m.timers))
	for name := range m.timers {
		out = append(out, name)
	}
	return out
}

// Remaining returns the time left on the named timer.
//
// If the manager has not started, the full duration is returned.
// If the timer has elapsed, zero is returned rather than a negative value.
// ErrUnknownTimer is returned if the name was never added.
func (m *Manager) Remaining(name string) (time.Duration, error) {
	d, ok := m.timers[name]
	if !ok {
		return 0, ErrUnknownTimer
	}
	if !m.started {
		return d, nil
	}
	elapsed := m.clock().Sub(m.start)
	if elapsed < 0 {
		// A clock that moves backwards should not produce more time than
		// the timer started with; clamp to the original duration.
		return d, nil
	}
	remaining := d - elapsed
	if remaining < 0 {
		return 0, nil
	}
	return remaining, nil
}

// Expired reports whether the named timer has elapsed.
//
// A timer that has not been started is never expired.
// ErrUnknownTimer is returned if the name was never added.
func (m *Manager) Expired(name string) (bool, error) {
	if _, ok := m.timers[name]; !ok {
		return false, ErrUnknownTimer
	}
	if !m.started {
		return false, nil
	}
	remaining, err := m.Remaining(name)
	if err != nil {
		return false, err
	}
	return remaining == 0, nil
}
