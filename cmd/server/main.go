package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync"
)

const (
	protocol     = "tcp"
	genericCrash = 1
)

type connection map[string]chan []byte

type application struct {
	port        int
	logger      *slog.Logger
	connections connection
	mu          sync.RWMutex
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	app := &application{
		logger: logger,
	}

	flag.IntVar(&app.port, "port", 8080, "Chat Server Port")
	flag.Parse()

	ln, err := net.Listen(protocol, fmt.Sprintf(":%d", app.port))
	if err != nil {
		app.logger.Error("unable to listen:", "err", err)
		os.Exit(genericCrash)
	}
	defer ln.Close()
	app.logger.Info("Chat Server Started and Listening", "port", app.port)
}
