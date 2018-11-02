package main

import (
	"code.cloudfoundry.org/cfdev/cfdevd/cmd"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"time"

	"code.cloudfoundry.org/cfdev/daemon"

	"github.com/spf13/cobra"
)

var (
	timesyncSocket = ""
	sockName       = "ListenSocket"
	doneChan       = make(chan bool, 10)
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			err := install(os.Args[0], os.Args[2:])
			if err != nil {
				log.Printf("Error: %s\n", err)
				os.Exit(1)
			}
		default:
			rootCmd := root()
			rootCmd.SetArgs(os.Args[1:])
			rootCmd.Execute()
		}
	}
}

func root() *cobra.Command {
	root := &cobra.Command{Use: "cfdevd"}
	root.PersistentFlags().StringVarP(&timesyncSocket, "timesyncSock", "t", "", "path to socket where host-timesync-daemon is listening")
	root.Run = func(_ *cobra.Command, _ []string) {
		log.Printf("Running cfdevd with timesyncSocket=%s\n", timesyncSocket)

		go registerSignalHandler()
		go syncTime(timesyncSocket)
		listenAndServe()
	}

	return root
}

func listenAndServe() {
	listeners, err := daemon.Listeners(sockName)
	if err != nil || len(listeners) != 1 {
		log.Fatalf("Failed to obtain socket from launchd: %s\n", err)
	}

	listener, ok := listeners[0].(*net.UnixListener)
	if !ok {
		log.Fatal("Failed to cast listener to unix listener")
	}

	for {
		select {
		case <-doneChan:
			log.Println("Terminating server listener...")
			return
		default:
			conn, err := listener.AcceptUnix()
			if err != nil {
				continue
			}

			if err := doHandshake(conn); err != nil {
				log.Printf("Handshake Error: %s\n", err)
				continue
			}

			command, err := cmd.UnmarshalCommand(conn)
			if err != nil {
				log.Printf("Command Error: %s\n", err)
				continue
			}

			command.Execute(conn)
			conn.Close()
		}
	}

}

func syncTime(socket string) {
	if socket == "" {
		return
	}

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-doneChan:
			log.Println("Terminating time sync...")
			return
		case <-ticker.C:
			// Only try to sync when the socket finally appears
			// to avoid the race condition
			if _, err := os.Stat(socket); os.IsNotExist(err) {
				continue
			}

			conn, err := net.DialUnix("unix", nil, &net.UnixAddr{
				Net:  "unix",
				Name: socket,
			})

			if err != nil {
				log.Printf("Timesync Error: %s\n", err)
				continue
			}

			// Only close if the error is nil
			// and thus 'conn' is not nil or it will panic
			conn.Close()
		}
	}
}

func registerSignalHandler() {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)

	<-sigc
	log.Println("Shutting down...")

	// We have at least 2 go functions running so
	// we send multiple signals to end them all
	doneChan <- true
	doneChan <- true
	doneChan <- true
}
