package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	sse "github.com/emmrk/sse/v2"
)

func TestReplayEventExpiry() {
	server := sse.New()
	server.AutoReplay = true
	server.EventTTL = 100 * time.Millisecond

	server.CreateStream("messages")

	mux := http.NewServeMux()
	mux.HandleFunc("/events", server.ServeHTTP)

	socket, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	// this server posts an event every millisecond and is configured to
	// retain events for 100 milliseconds therefore a client that connects
	// "later" will get a replay of about 100 events
	go func() {
		ticker := time.NewTicker(time.Millisecond)
		for {
			<-ticker.C
			server.Publish("messages", &sse.Event{
				Data: []byte("ping"),
			})
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	// this client subscribes more or less immediately and so receives all
	// the published events
	earlyClientCount := 0
	go func(count *int) {
		client := sse.NewClient("http://localhost:8080/events")
		client.SubscribeWithContext(ctx, "messages", func(msg *sse.Event) {
			*count++
		})
		done <- struct{}{}
	}(&earlyClientCount)

	// this client actually matters for the integration test: it joins late
	// and should receive 100 events from the replay before catching up
	// with the stream instead of the 300 it skipped
	lateClientCount := 0
	go func(count *int) {
		time.Sleep(300 * time.Millisecond)

		client := sse.NewClient("http://localhost:8080/events")
		client.SubscribeWithContext(ctx, "messages", func(msg *sse.Event) {
			*count++
		})
		done <- struct{}{}
	}(&lateClientCount)

	// let this run for a second and then compare the counts of events
	// received: the late client should receive about 200 events less than
	// the early one (it joined at 300 events in, but only 100 will be
	// replayed)
	time.AfterFunc(time.Second, func() {
		// stop clients to get the final counts, make sure that they
		// are down before continuing to avoid a data race (not that we
		// care about it much here)
		cancel()
		<-done
		<-done

		if earlyClientCount-lateClientCount < 250 {
			os.Exit(0)
		}

		fmt.Printf("Early client got %d events, late client got %d events. The difference is %d, this is not right\n", earlyClientCount, lateClientCount, earlyClientCount-lateClientCount)
		os.Exit(1)
	})

	http.Serve(socket, mux)
}
