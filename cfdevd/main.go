package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"io"

	"time"

	"code.cloudfoundry.org/cfdev/cfdevd/cmd"
	"code.cloudfoundry.org/cfdev/daemon"

	"github.com/spf13/cobra"
)

const SockName = "ListenSocket"

func handleRequest(conn *net.UnixConn) {
	if err := doHandshake(conn); err != nil {
		fmt.Println("Handshake Error: ", err)
		return
	}

	command, err := cmd.UnmarshalCommand(conn)
	if err != nil {
		fmt.Println("Command:", err)
		return
	}
	command.Execute(conn)
}

func registerSignalHandler() {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func(c chan os.Signal) {
		sig := <-c
		log.Printf("Caught signal %s: shutting down.", sig)
		os.Exit(0)
	}(sigc)
}

func install(programSrc string, args []string) {
	lctl := daemon.New("")
	program := "/Library/PrivilegedHelperTools/org.cloudfoundry.cfdevd"
	programArgs := []string{program}
	programArgs = append(programArgs, args...)
	cfdevdSpec := daemon.DaemonSpec{
		Label:            "org.cloudfoundry.cfdevd",
		Program:          program,
		ProgramArguments: programArgs,
		RunAtLoad:        false,
		Sockets: map[string]string{
			SockName: "/var/tmp/cfdevd.socket",
		},
		StdoutPath: "/var/tmp/cfdevd.stdout.log",
		StderrPath: "/var/tmp/cfdevd.stderr.log",
	}
	if err := copyExecutable(programSrc, program); err != nil {
		fmt.Println("Failed to copy cfdevd: ", err)
	}
	if err := lctl.AddDaemon(cfdevdSpec); err != nil {
		fmt.Println("Failed to install cfdevd: ", err)
	}
}

func copyExecutable(src string, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	target, err := os.Create(dest)
	if err != nil {
		return err
	}

	if err = os.Chmod(dest, 0744); err != nil {
		return err
	}

	binData, err := os.Open(src)
	if err != nil {
		return err
	}

	_, err = io.Copy(target, binData)
	return err
}

func timesync(socket string) {
	for {
		fmt.Printf("dialing socket %s \n", socket)
		conn, _ := net.DialUnix("unix", nil, &net.UnixAddr{
			Net:  "unix",
			Name: socket,
		})
		if conn != nil {
			conn.CloseWrite()
		}
		time.Sleep(5 * time.Second)
	}
}

func run(timesyncSocket string) {
	fmt.Println("SOCKET: ", timesyncSocket)
	if timesyncSocket != "" {
		fmt.Println("timesync")
		go timesync(timesyncSocket)
	}
	registerSignalHandler()
	listeners, err := daemon.Listeners(SockName)
	if err != nil || len(listeners) != 1 {
		log.Fatal("Failed to obtain socket from launchd")
	}
	listener, ok := listeners[0].(*net.UnixListener)
	if !ok {
		log.Fatal("Failed to cast listener to unix listener")
	}
	for {
		conn, err := listener.AcceptUnix()
		if err != nil {
			continue
		}
		defer conn.Close()
		go handleRequest(conn)
	}
}

func root() *cobra.Command {
	root := &cobra.Command{Use: "cfdevd"}
	flags := root.PersistentFlags()
	var timesyncSocket string
	flags.StringVarP(&timesyncSocket, "timesyncSock", "t", "", "path to socket where host-timesync-daemon is listening")
	root.Run = func(_ *cobra.Command, _ []string) {
		log.Printf("running cfdevd with timesyncSocket=%s", timesyncSocket)
		run(timesyncSocket)
	}
	return root
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			install(os.Args[0], os.Args[2:])
		default:
			rootCmd := root()
			rootCmd.SetArgs(os.Args[1:])
			rootCmd.Execute()
		}
	}
}
