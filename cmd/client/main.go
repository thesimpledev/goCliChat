package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/thesimpledev/gochat/internal/protocol"
)

type application struct {
	ctx       context.Context
	logger    *slog.Logger
	connected bool
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	app := &application{
		ctx:    ctx,
		logger: logger,
	}
	parts := strings.Split(os.Args[1], "@")
	userName := parts[0]
	address := parts[1]
	register := os.Args[2] == "-r"

	conn, err := net.Dial(protocol.Protocol, address)
	if err != nil {
		app.logger.Error("unable to connect to server", "err", err)
	}
	defer conn.Close()
	tcpConn := conn.(*net.TCPConn)
	tcpConn.SetKeepAlive(true)
	tcpConn.SetKeepAlivePeriod(protocol.KeepAliveTimer)

	app.logger.Info("Connected to Chat Server", "address", address)

	go app.readServerMessages(ctx, conn)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		message := scanner.Text()
		message += "\n"
		_, err := conn.Write([]byte(message))
		if err != nil {
			log.Fatal("Failed to send message:", err)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal("Error reading input:", err)
	}

}

func (app *application) readServerMessages(ctx context.Context, conn net.Conn) {
	buf := make([]byte, 1024)
	for {
		_, err := io.ReadFull(conn, buf)
		if err != nil {
			app.logger.Error("disconnected from server", "err", err)
			os.Exit(protocol.ErrGenericCrash)
		}
		payload := app.decrypt(buf)
		frame, err := protocol.UnpackFrame(payload)
		if err != nil {
			app.logger.Error("unable to unpack frame", "err", err)
		}
		chatMessage := fmt.Sprintf("%s: %s", frame.UserName, frame.Message)
		fmt.Print(chatMessage)

	}
}

func (app *application) decrypt(buf []byte) []byte {
	return buf
}
