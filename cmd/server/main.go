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
	userName  string
	channel   chan []byte
	publicKey string
}

type connections map[string]*connection

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
	var connection *connection
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
			if connection != nil {
				leaveMessage := fmt.Appendf(nil, "*** %s has left the chat", connection.userName)
				app.broadcast(connection.userName, leaveMessage)
				app.mu.Lock()
				close(connection.channel)
				app.mu.Unlock()
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
			userName := string(message.userName)
			connection = app.registerConnection(userName)
			joinMessage := fmt.Appendf(nil, "*** %s joined the chat\n", message.userName)
			app.broadcast(userName, joinMessage)
			go app.writeOutgoing(connection, conn)
		case protocol.CMD_REGISTER:
			// TODO: handler registeration
		case protocol.CMD_CHAT:
			if connection == nil {
				app.logger.Error("chat attempted before connection established")
				continue
			}
			chatMessage := fmt.Sprintf("%s: %s", message.userName, message.message)
			app.broadcast(connection.userName, []byte(chatMessage))
		case protocol.CMD_ROTATE_KEY:
			if connection == nil {
				app.logger.Error("chat attempted before connection established")
				continue
			}
			// TODO: Handle Key Rotation/clos
		case protocol.CMD_DELETE_ME:
			if connection == nil {
				app.logger.Error("chat attempted before connection established")
				continue
			}

			// TODO: Handle Account Deletion
		default:
			app.logger.Error("unknown command", "command", command)
		}
	}

}

func (app *application) writeOutgoing(connection *connection, conn net.Conn) {
	for msg := range connection.channel {
		_, err := conn.Write(msg)
		if err != nil {
			app.logger.Error("write error", "err", err)
			return
		}
	}
}

func (app *application) broadcast(sender string, message []byte) {
	app.mu.RLock()
	defer app.mu.RUnlock()

	for _, conn := range app.connections {
		if conn.userName == sender {
			continue
		}

		app.queueMessage(conn.channel, conn.userName, message)

	}
}

func (app *application) queueMessage(channel chan []byte, userName string, msg []byte) {
	select {
	case channel <- msg:
	default:
		app.logger.Error("unable to send message to channel", "channel", userName)
	}
}

func (app *application) decryptFrame(frame []byte) ([]byte, error) {
	// TODO: decrypt logic
	return frame, nil
}

func (app *application) unpackFrame(packedFrame []byte) (*messageContainer, error) {
	// TODO: Need to finish setting up the packedFrame parsing
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

func (app *application) registerConnection(userName string) *connection {
	connection := &connection{
		userName: userName,
		channel:  make(chan []byte, 10),
	}
	app.mu.Lock()
	defer app.mu.Unlock()

	if old, exists := app.connections[userName]; exists {
		close(old.channel)
	}
	app.connections[userName] = connection

	return connection

}
