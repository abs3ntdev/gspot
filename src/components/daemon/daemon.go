package daemon

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"time"

	"go.uber.org/fx"

	"github.com/abs3ntdev/gspot/src/components/commands"
	"github.com/abs3ntdev/gspot/src/config"
)

func Run(c *commands.Commander, conf *config.Config, s fx.Shutdowner) {
	for {
		err := startServer(c, conf)
		if err != nil {
			log.Printf("Server error: %v", err)
			time.Sleep(time.Second)
			continue
		}
		break
	}
}

func startServer(c *commands.Commander, conf *config.Config) error {
	socketPath := conf.SocketPath

	if _, err := os.Stat(socketPath); err == nil {
		if err := os.Remove(socketPath); err != nil {
			return fmt.Errorf("failed to remove existing socket: %w", err)
		}
	}

	commandHandler := &Handler{Commander: c}
	server := rpc.NewServer()
	if err := server.Register(commandHandler); err != nil {
		return fmt.Errorf("failed to register RPC handler: %w", err)
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("listen error: %w", err)
	}
	defer listener.Close()

	if err := os.Chmod(socketPath, 0o666); err != nil {
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}

	log.Println("Daemon is listening on", socketPath)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		codec := NewLoggingServerCodec(conn)
		go handleConnection(server, codec)
	}
}

func handleConnection(server *rpc.Server, codec rpc.ServerCodec) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered in handleConnection: %v", r)
		}
	}()
	server.ServeCodec(codec)
}
