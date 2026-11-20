# simple-countdown-timer-manager

Tracks multiple named countdown timers with a shared start instant and on-demand remaining-time queries.

```go
package main

import (
	"fmt"
	"time"

	countdown "simplecountdowntimermanager"
)

func main() {
	m := countdown.New(time.Now)
	m.Add("tea", 3*time.Minute)
	m.Add("eggs", 5*time.Minute)

	if err := m.Start(); err != nil {
		panic(err)
	}

	// Later, in your own loop:
	remaining, err := m.Remaining("tea")
	if err != nil {
		panic(err)
	}
	fmt.Println(remaining)
}
```

## Why this exists

A program that runs several countdowns usually wants them to start together and wants to ask "how long is left?" without storing a separate deadline per timer. Storing deadlines means every timer carries its own start instant, which drifts if starts are recorded at different moments. This manager stores one start instant and a duration per name, so all timers share the same origin by construction.

The trade-off is that timers cannot be started independently. If you need per-timer start times, this is the wrong shape; a map of `time.Time` deadlines would serve you better.

## The awkward edge

Remaining time is computed from the injected clock, not from wall time. If the clock moves backwards, `Remaining` clamps to the original duration rather than reporting more time than the timer started with. If a timer has elapsed, `Remaining` returns zero rather than a negative duration, and `Expired` reports true. Before `Start` is called, `Remaining` returns the full duration and `Expired` is false.

## API

- `New(clock func() time.Time) *Manager` — construct with `time.Now` or a fake clock.
- `(*Manager) Add(name string, d time.Duration)` — register or replace a timer; negative durations are clamped to zero.
- `(*Manager) Start() error` — record the shared start instant; returns `ErrAlreadyStarted` on a second call.
- `(*Manager) Started() bool` — whether `Start` has succeeded.
- `(*Manager) Names() []string` — registered names, as a fresh copy.
- `(*Manager) Remaining(name string) (time.Duration, error)` — time left, clamped at zero; returns `ErrUnknownTimer` for unknown names.
- `(*Manager) Expired(name string) (bool, error)` — whether the timer has elapsed; returns `ErrUnknownTimer` for unknown names.
