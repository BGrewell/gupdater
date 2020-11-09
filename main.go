package main

import (
	"fmt"
	"github.com/BGrewell/gupdater/autoupdater"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	updater := autoupdater.AutoUpdater{}
	apps, err := updater.ParseConfiguration("example.yaml")
	if err != nil {
		log.Fatalf("failed to parse configuration: %v", err)
	}

	running := true
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL) //TODO use one of the signals to trigger an update

	go func() {
		sig := <-sigs
		fmt.Printf("got a signal to quit: %v\n", sig)
		running = false
	}()

	nextCheck := time.Now().Unix()
	for running {

		if time.Now().Unix() >= nextCheck {
			log.Println("[+] checking for updates")
			for _, app := range(apps) {
				err = updater.Update(app)
				if err != nil {
					log.Printf(" %s\n", err)
					os.Exit(-1)
				}
			}
			nextCheck = time.Now().Unix() + 5
		}
		time.Sleep(1 * time.Second)
	}

}
