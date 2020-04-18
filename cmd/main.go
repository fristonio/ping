package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/fristonio/ping/pkg/ping"
	log "github.com/sirupsen/logrus"
)

var (
	// verbose if true then debug information is also logged to stdout
	verbose bool

	// Host is the host to ping
	Host string
)

func main() {
	log.Debug("ping-go is a go implementation of linux ping utility")
	if Host == "" {
		log.Error("host is a required parameter")
		os.Exit(1)
	}

	log.Debugf("provided host for ping is: %s", Host)

	pinger, err := ping.NewPinger(Host)
	if err != nil {
		log.Errorf("error while initializing pinger instance: %s", err)
		os.Exit(1)
	}

	// Set up signal handlers for pinger running.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		log.Info("Signal recieved, shutting down pinger instance.")
		pinger.Shutdown()
		pinger.PrintStats()
		os.Exit(0)
	}()

	err = pinger.Run()
	if err != nil {
		log.Errorf("error while running pinger: %s", err)
	}
}

func init() {
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	flag.StringVar(&Host, "host", "", "host to ping.")
	flag.StringVar(&Host, "h", "", "host to ping (shorthand)")

	flag.BoolVar(&verbose, "verbose", false, "print verbose information to stdout")
	flag.BoolVar(&verbose, "v", false, "print verbose information to stdout (shorthand)")

	flag.Parse()
	if verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}
