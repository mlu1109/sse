# sseclt

sseclt (sse client) is a server sent event client library for golang.

## Usage

```go
package main

import (
	"context"
	"log"

	"github.com/mlu1109/sseclt"
)

func main() {
	url := "..."
	r, err := sseclt.
		NewRequest(url). // This is merely a convenience method, r is just an *http.Response
		Do()
	if err != nil {
		log.Panic(err)
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	err = sseclt.
		NewStream(ctx).
		OnOpen(onOpen).
		OnClose(onClose).
		OnError(onError).
		OnEvent(onEvent).
		Subscribe(r)
	if err != nil {
		log.Panic(err)
	}

	// ...

	cancelFunc() // Will stop the subscription and fire the onClose handler
}

func onOpen() {
	log.Printf("🚀 First event received!")
}

func onClose(err error) {
	log.Printf("📪 Subscription closed: %s", err)
}

func onError(err error) {
	log.Printf("💥 Encountered error: %s", err)
}

func onEvent(event sseclt.Event) {
	log.Printf("📬 Received event: %s", event)
}

```