package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/moby/vpnkit/go/pkg/libproxy"
)

func main() {
	if err := exportPorts(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func exportPorts() error {
	var (
		ip        = flag.String("container-ip", "", "container ip")
		startPort = flag.Int("start-port", -1, "start port")
		endPort   = flag.Int("end-port", -1, "end port")
	)

	flag.Parse()

	for port := *startPort; port < *endPort+1; port++ {
		host := &net.TCPAddr{IP: net.ParseIP(*ip), Port: port}
		container := &net.TCPAddr{IP: net.ParseIP(*ip), Port: port}

		ctl, err := libproxy.ExposePort(host, container)

		if err != nil {
			return fmt.Errorf("unable to expose port: %s", err)
		}

		defer ctl.Close()
	}

	log.Println("Proxy running")

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSTOP)

	<-s
	return nil
}
