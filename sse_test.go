package sseclt_test

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
)

type Event struct {
	id      string
	data    string
	event   string
	comment string
}

func (e Event) String() string {
	str := ""
	if e.id != "" {
		str += fmt.Sprintf("id: %s\n", e.id)
	}
	if e.event != "" {
		str += fmt.Sprintf("event: %s\n", e.event)
	}
	if e.data != "" {
		str += fmt.Sprintf("data: %s\n", e.data)
	}
	if e.comment != "" {
		str += fmt.Sprint(':', e.comment, '\n')
	}
	if str == "" {
		return str
	} else {
		return str + "\n"
	}
}

func NewTestSSEServer(events <-chan fmt.Stringer) *httptest.Server {
	mux := http.NewServeMux()
	clients := make(map[int]chan<- fmt.Stringer, 0)
	go func() {
		for {
			e, ok := <-events
			if !ok {
				return
			}
			for _, client := range clients {
				client <- e
			}
		}
	}()
	mux.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		clientID := len(clients)
		flusher, ok := rw.(http.Flusher)
		if !ok {
			http.Error(rw, "SSE not supported", http.StatusInternalServerError)
		}
		rw.Header().Set("Access-Control-Allow-Origin", "*")
		rw.Header().Set("Access-Control-Allow-Headers", "*")
		rw.Header().Set("Content-Type", "text/event-stream")
		rw.Header().Set("Cache-Control", "no-cache")
		rw.Header().Set("Connection", "keep-alive")
		done := r.Context().Done()
		ch := make(chan fmt.Stringer)
		clients[clientID] = ch
		fmt.Fprint(rw, ": Connected\n\n")
		flusher.Flush()
		for {
			select {
			case <-done:
				delete(clients, clientID)
				close(ch)
				return
			case e, ok := <-ch:
				if ok {
					str := e.String()
					_, err := fmt.Fprint(rw, str)
					if err != nil {
						log.Panic(err)
					}
					flusher.Flush()
				} else {
					return
				}
			}
		}
	})
	return httptest.NewServer(mux)
}
