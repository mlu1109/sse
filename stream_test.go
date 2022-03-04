package sseclt_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mlu1109/sseclt"
	"github.com/stretchr/testify/assert"
)

func TestSubscribe(t *testing.T) {
	setupTestServerAndSubscription := func(
		ctx context.Context,
		eventsToPublish <-chan fmt.Stringer,
		onEvent func(event sseclt.Event),
		onError func(err error),
		onOpen func(),
		onClose func(err error),
	) (*sseclt.Stream, error) {
		srv := NewTestSSEServer(eventsToPublish)
		url := srv.URL
		response, err := sseclt.
			NewRequest(url).
			Do()
		if err != nil {
			return nil, err
		}
		stream := sseclt.
			NewStream(ctx).
			OnOpen(onOpen).
			OnClose(onClose).
			OnEvent(onEvent).
			OnError(onError)
		err = stream.Subscribe(response)
		if err != nil {
			return nil, err
		}
		return stream, nil
	}

	t.Run("When events are received, then call onEvent with the expected events", func(t *testing.T) {
		// Given
		eventGenerator := func(ctx context.Context) (func(), <-chan fmt.Stringer) {
			ch := make(chan fmt.Stringer)
			routine := func() {
				for i := 0; ; i++ {
					select {
					case ch <- Event{
						id:    fmt.Sprint("id", i),
						data:  fmt.Sprint("data", i),
						event: fmt.Sprint("event", i)}:
					case <-ctx.Done():
						return
					}
				}
			}
			return routine, ch
		}
		ctx, cancelFn := context.WithCancel(context.Background())
		generator, eventsToPublish := eventGenerator(ctx)
		active := true
		onClose := func(err error) {
			active = false
		}
		var received []sseclt.Event
		onEvent := func(event sseclt.Event) {
			received = append(received, event)
		}
		// When
		setupTestServerAndSubscription(ctx, eventsToPublish, onEvent, nil, nil, onClose)
		go generator()
		numberOfEventsBeforeCancel := 10
		for active {
			if len(received) >= numberOfEventsBeforeCancel {
				cancelFn()
			}
		}
		cancelFn()
		// Then ...
		assert.Len(t, received, numberOfEventsBeforeCancel) // ... the expected number of events was received
		for i, event := range received {                    // ... the expected events were received in the correct order
			expectedData := fmt.Sprint("data", i)
			expectedEvent := fmt.Sprint("event", i)
			expectedID := fmt.Sprint("id", i)
			assert.Equal(t, expectedData, event["data"])
			assert.Equal(t, expectedEvent, event["event"])
			assert.Equal(t, expectedID, event["id"])
		}
	})

	t.Run("When the first event is published, then call OnOpen before OnEvent", func(t *testing.T) {
		// Given
		chevent := make(chan fmt.Stringer)
		onOpenCalled := false
		onEventCalled := false
		onOpen := func() {
			assert.False(t, onOpenCalled)
			assert.False(t, onEventCalled)
			onOpenCalled = true
		}
		onEvent := func(event sseclt.Event) {
			onEventCalled = true
		}
		ctx, cancelFn := context.WithCancel(context.Background())
		_, err := setupTestServerAndSubscription(ctx, chevent, onEvent, nil, onOpen, nil)
		assert.Nil(t, err)
		// When
		chevent <- Event{id: "test"}
		time.Sleep(1 * time.Millisecond)
		cancelFn()
		time.Sleep(1 * time.Millisecond)
		// Then (order is validated in the onOpen function above)
		assert.True(t, onOpenCalled)
		assert.True(t, onEventCalled)
	})

	t.Run("When cancelFn is called, then call OnClose with the expected error", func(t *testing.T) {
		// Given
		chevent := make(chan fmt.Stringer)
		var actualError error
		onCloseCalled := false
		onClose := func(err error) {
			actualError = err
			onCloseCalled = true
		}
		ctx, cancelFn := context.WithCancel(context.Background())
		_, err := setupTestServerAndSubscription(ctx, chevent, nil, nil, nil, onClose)
		assert.Nil(t, err)
		// When
		cancelFn()
		time.Sleep(1 * time.Millisecond)
		// Then ...
		assert.True(t, onCloseCalled)
		assert.ErrorIs(t, actualError, context.Canceled)
	})

	t.Run("When the subscription body reader encounters an error, then call OnError with the expected error", func(t *testing.T) {
		t.Skip("TODO")
	})

	t.Run("When Subscribe is called after the context is closed, then return ctx.Error()", func(t *testing.T) {
		// Given
		chevent := make(chan fmt.Stringer)
		srv := NewTestSSEServer(chevent)
		url := srv.URL
		response, err := sseclt.
			NewRequest(url).
			Do()
		assert.Nil(t, err)
		ctx, cancelFn := context.WithCancel(context.Background())
		stream := sseclt.NewStream(ctx)
		err = stream.Subscribe(response)
		assert.Nil(t, err)
		// When
		cancelFn()
		time.Sleep(1 * time.Millisecond)
		actualError := stream.Subscribe(response)
		// Then ...
		assert.ErrorIs(t, actualError, context.Canceled)
	})
}
