package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/thesimpledev/gochat/internal/protocol"
)

type connection struct {
	channel   chan []byte
	publicKey string
}

type connections map[string]connection

type messageContainer struct {
	userName  []byte
	signature []byte
	message   []byte
}

type application struct {
	ctx         context.Context
	port        int
	logger      *slog.Logger
	connections connections
	mu          sync.RWMutex
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	app := &application{
		ctx:         ctx,
		logger:      logger,
		connections: make(connections),
	}

	flag.IntVar(&app.port, "port", protocol.DefaultPort, "Chat Server Port")
	flag.Parse()

	ln, err := net.Listen(protocol.Protocol, fmt.Sprintf(":%d", app.port))
	if err != nil {
		app.logger.Error("unable to listen:", "err", err)
		os.Exit(protocol.ErrGenericCrash)
	}
	defer ln.Close()
	app.logger.Info("Chat Server Started and Listening", "port", app.port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			app.logger.Error("failed to accept connection", "err", err)
			continue
		}

		go app.handleConnection(conn)
	}
}

func (app *application) handleConnection(conn net.Conn) {
	var userName string
	defer conn.Close()
	tcpConn := conn.(*net.TCPConn)
	tcpConn.SetKeepAlive(true)
	tcpConn.SetKeepAlivePeriod(protocol.KeepAliveTimer)

	buf := make([]byte, protocol.PayloadTotal)

	for {
		_, err := io.ReadFull(conn, buf)
		if err != nil {
			if err != io.EOF {
				app.logger.Error("read error", "err", err)
			}
			if userName != "" {
				leaveMessage := fmt.Appendf(nil, "*** %s has left the chat", userName)
				app.broadcast(leaveMessage)
				close(app.connections[userName].channel)
			}
			return
		}

		command := buf[0]

		if command == protocol.CMD_KEY_REQ {
			//TODO: Need to return server key
			continue
		}

		frame, err := app.decryptFrame(buf[1:])
		if err != nil {
			app.logger.Error("decryption of frame failed", "err", err)
			continue
		}

		message, err := app.unpackFrame(frame)
		if err != nil {
			app.logger.Error("failedto unpack frame", "err", err)
			continue
		}

		switch command {
		case protocol.CMD_CONNECTION:
			userName = string(message.userName)
			app.registerConnection(userName)
			joinMessage := fmt.Appendf(nil, "*** %s joined the chat\n", message.userName)
			app.broadcast(joinMessage)
		case protocol.CMD_REGISTER:
			// TODO: handler registeration
		case protocol.CMD_CHAT:
			chatMessage := fmt.Sprintf("%s: %s", message.userName, message.message)
			app.broadcast([]byte(chatMessage))
		case protocol.CMD_ROTATE_KEY:
			// TODO: Handle Key Rotation
		case protocol.CMD_DELETE_ME:
			// TODO: Handle Account Deletion
		default:
			app.logger.Error("unknown command", "command", command)
		}
	}

}

func (app *application) decryptFrame(frame []byte) ([]byte, error) {
	return frame, nil
}

func (app *application) unpackFrame(packedFrame []byte) (*messageContainer, error) {
	userName := packedFrame[:]
	message := packedFrame[:]
	signature := packedFrame[:]
	mc := &messageContainer{
		userName:  userName,
		message:   message,
		signature: signature,
	}
	return mc, nil
}

func (app *application) broadcast(message []byte) {

}
func (app *application) registerConnection(userName string) {
	connection := &connection{
		channel: make(chan []byte),
	}
	app.mu.Lock()
	defer app.mu.Unlock()

	if old, exists := app.connections[userName]; exists {
		close(old.channel)
	}
	app.connections[userName] = *connection

}
