package ui

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
)

type Spinner struct {
	message string
	ctx     context.Context
	cancel  context.CancelFunc
	done    chan bool
}

var spinnerChars = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func NewSpinner(message string) *Spinner {
	ctx, cancel := context.WithCancel(context.Background())
	sp := &Spinner{
		message: message,
		ctx:     ctx,
		cancel:  cancel,
		done:    make(chan bool, 1),
	}

	go sp.run()
	return sp
}

func (s *Spinner) run() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	i := 0
	for {
		select {
		case <-s.ctx.Done():
			fmt.Fprint(os.Stderr, "\r\033[K")
			s.done <- true
			return
		case <-ticker.C:
			char := spinnerChars[i%len(spinnerChars)]
			fmt.Fprintf(os.Stderr, "\r%s %s", color.CyanString(char), s.message)
			i++
		}
	}
}

func (s *Spinner) Stop() {
	s.cancel()
	<-s.done
}

func ShowSpinner(message string, fn func() error) error {
	sp := NewSpinner(message)
	defer sp.Stop()
	return fn()
}

