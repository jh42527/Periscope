package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// TO-DO: Better Error Handling
func main() {
	// get config path flag or default
	configPtr := flag.String("c", "config.yml", "config to process")

	flag.Parse()

	// read config yml
	config, err := readConfig(*configPtr)
	if err != nil {
		processError(err, true, config)
	}

	// api gateway
	g := Gateway{}

	g.Start(config)

	// listen for kill signal
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	run := true

	for run == true {
		select {
		case sig := <-sigchan:
			fmt.Println("")
			fmt.Println("Terminating,", sig)
			run = false

		default:
		}
	}
}

func processError(err error, fatal bool, config *Config) {
	fmt.Println(err)

	if fatal {
		os.Exit(1)
	}
}
