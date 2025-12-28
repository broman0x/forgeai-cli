package ui

import (
	"fmt"
	"sync"
	"time"
)

type Spinner struct {
	stop    chan bool
	wg      sync.WaitGroup
	message string
}

func NewSpinner(msg string) *Spinner {
	return &Spinner{
		stop:    make(chan bool),
		message: msg,
	}
}

func (s *Spinner) Start() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
		for {
			select {
			case <-s.stop:
				fmt.Printf("\r\033[K")
				return
			default:
				fmt.Printf("\r\033[36m%s\033[0m %s ", frames[i%len(frames)], s.message)
				i++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

func (s *Spinner) Stop() {
	s.stop <- true
	s.wg.Wait()
}