package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	mode := flag.String("mode", "replay", "Operating mode: record or replay")
	port := flag.String("port", "8765", "Port to listen on")
	target := flag.String("target", "", "Target API URL (required in record mode)")
	cassettes := flag.String("cassettes", "e2e/cassettes/happypath", "Directory for cassette files")
	flag.Parse()

	if *mode != "record" && *mode != "replay" {
		fmt.Fprintf(os.Stderr, "Error: --mode must be 'record' or 'replay', got %q\n", *mode)
		os.Exit(1)
	}

	if *mode == "record" && *target == "" {
		fmt.Fprintln(os.Stderr, "Error: --target is required in record mode")
		os.Exit(1)
	}

	// Ensure cassettes directory exists.
	if err := os.MkdirAll(*cassettes, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not create cassettes dir %q: %v\n", *cassettes, err)
		os.Exit(1)
	}

	var handler http.Handler
	var recorder *recordProxy

	switch *mode {
	case "replay":
		proxy, err := newReplayProxy(*cassettes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initialising replay proxy: %v\n", err)
			os.Exit(1)
		}
		handler = proxy
		fmt.Printf("Replay proxy ready on :%s (cassettes: %s)\n", *port, *cassettes)

	case "record":
		proxy, err := newRecordProxy(*target, *cassettes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initialising record proxy: %v\n", err)
			os.Exit(1)
		}
		recorder = proxy
		handler = proxy
		fmt.Printf("Record proxy ready on :%s → %s (cassettes: %s)\n", *port, *target, *cassettes)
	}

	srv := &http.Server{
		Addr:    ":" + *port,
		Handler: handler,
	}

	// Graceful shutdown on SIGINT/SIGTERM.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	}()

	<-stop
	fmt.Println("\nShutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)

	// Save cassette if we were recording.
	if recorder != nil {
		if err := recorder.save(); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving cassette: %v\n", err)
			os.Exit(1)
		}
	}
}
