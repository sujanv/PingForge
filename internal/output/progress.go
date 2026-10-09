package output

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
)

// Spinner shows animated terminal feedback during long network operations.
type Spinner struct {
	mu     sync.Mutex
	msg    string
	w      io.Writer
	stop   chan struct{}
	active bool
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// NewSpinner creates a new terminal spinner.
func NewSpinner(msg string) *Spinner {
	return &Spinner{
		msg:  msg,
		w:    os.Stderr,
		stop: make(chan struct{}),
	}
}

// Start displays the spinning animation.
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.active {
		s.mu.Unlock()
		return
	}
	s.active = true
	s.mu.Unlock()

	go func() {
		idx := 0
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-s.stop:
				fmt.Fprintf(s.w, "\r\033[K")
				return
			case <-ticker.C:
				s.mu.Lock()
				frame := color.CyanString(spinnerFrames[idx%len(spinnerFrames)])
				fmt.Fprintf(s.w, "\r %s %s", frame, s.msg)
				idx++
				s.mu.Unlock()
			}
		}
	}()
}

// Stop terminates the spinner and clears the line.
func (s *Spinner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.active {
		return
	}
	s.active = false
	close(s.stop)
	time.Sleep(50 * time.Millisecond)
}
