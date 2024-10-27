package daemon

import (
	"log"
	"net"
	"net/rpc"
	"os"

	"go.uber.org/fx"

	"git.asdf.cafe/abs3nt/gspot/src/components/commands"
	"git.asdf.cafe/abs3nt/gspot/src/config"
)

func Run(c *commands.Commander, conf *config.Config, s fx.Shutdowner) {
	socketPath := conf.SocketPath
	if _, err := os.Stat(socketPath); err == nil {
		os.Remove(socketPath)
	}

	CommandHandler := Handler{
		Commander: c,
	}

	server := rpc.NewServer()
	server.Register(&CommandHandler)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatal("Listen error:", err)
	}
	defer listener.Close()
	os.Chmod(socketPath, 0o666)

	log.Println("Daemon is listening on", socketPath)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		codec := NewLoggingServerCodec(conn)
		go server.ServeCodec(codec)
	}
}
