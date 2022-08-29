package sseclt

import (
	"bufio"
	"context"
	"net/http"
	"strings"
)

type Status int

type Stream struct {
	ctx     context.Context
	onEvent func(event Event)
	onError func(err error)
	onOpen  func()
	onClose func(err error)
	closed  bool
}

func NewStream(ctx context.Context) *Stream {
	return &Stream{
		ctx:     ctx,
		onEvent: func(event Event) {},
		onError: func(err error) {},
		onOpen:  func() {},
		onClose: func(err error) {},
		closed:  false,
	}
}

// OnOpen is called when the first event is received. OnOpen is called before OnEvent.
func (s *Stream) OnOpen(onOpen func()) *Stream {
	if onOpen == nil {
		s.onOpen = func() {}
	} else {
		s.onOpen = onOpen
	}
	return s
}

// OnError is called when the body reader encounters an error.
// OnError will not mark the context as done.
// Make sure to cancel the context if you want to stop the goroutine that handles the subscription.
func (s *Stream) OnError(onError func(err error)) *Stream {
	if onError == nil {
		s.onError = func(err error) {}
	} else {
		s.onError = onError
	}
	return s
}

// OnClose is called when the context is done.
// error will be the same value that was returned by s.ctx.Error()
func (s *Stream) OnClose(onClose func(err error)) *Stream {
	if onClose == nil {
		s.onClose = func(err error) {}
	} else {
		s.onClose = onClose
	}
	return s
}

// OnEvent is called when a new message is received.
func (s *Stream) OnEvent(onEvent func(event Event)) *Stream {
	if onEvent == nil {
		s.onEvent = func(event Event) {}
	} else {
		s.onEvent = onEvent
	}
	return s
}

// Subscribe starts a subscription that is canceled when the context is done.
func (s *Stream) Subscribe(resp *http.Response) error {
	err := s.ctx.Err()
	if err != nil {
		return err
	}

	go func() {
		defer func() {
			resp.Body.Close()
		}()
		chevent, cherr := s.reader(resp)
		connected := false
		for {
			select {
			case ev := <-chevent:
				if !connected {
					s.onOpen()
					connected = true
				}
				s.onEvent(ev)
			case err := <-cherr:
				s.onError(err)
			case <-s.ctx.Done():
				s.closed = true
				err := s.ctx.Err()
				s.onClose(err)
				return
			}
		}
	}()

	return nil
}

func (s *Stream) reader(resp *http.Response) (<-chan Event, <-chan error) {
	success := make(chan Event)
	errch := make(chan error)

	go func() {
		event := make(Event)
		reader := bufio.NewReader(resp.Body)
		for !s.closed {
			bytes, err := reader.ReadBytes('\n')
			line := strings.TrimSuffix(string(bytes), "\n")
			line = strings.TrimSuffix(line, "\r")
			if err != nil {
				errch <- err
			} else {
				if len(line) != 0 {
					event.AddLine(string(line))
				} else {
					success <- event
					event = make(Event)
				}
			}
		}
	}()

	return success, errch
}
