package daemon

import (
	"log"
	"net"
	"net/rpc"
	"os"
	"time"

	"go.uber.org/fx"

	"git.asdf.cafe/abs3nt/gspot/src/components/commands"
	"git.asdf.cafe/abs3nt/gspot/src/config"
)

func Run(c *commands.Commander, conf *config.Config, s fx.Shutdowner) {
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Recovered in Run: %v", r)
				}
			}()

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
				log.Println("Listen error:", err)
				time.Sleep(time.Second)
				return
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
				go func() {
					defer func() {
						if r := recover(); r != nil {
							log.Printf("Recovered in ServeCodec goroutine: %v", r)
						}
					}()
					server.ServeCodec(codec)
				}()
			}
		}()
		time.Sleep(time.Second)
	}
}
