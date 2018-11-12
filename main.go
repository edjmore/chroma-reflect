package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/edjmore/chroma-reflect/background"
	"github.com/edjmore/chroma-reflect/chroma"
)

func main() {
	cli := chroma.NewClient()
	cli.Register()
	defer cli.Unregister()

	// This goroutine listens for interrupts so we can unregister the app before exiting.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Printf("caught signal: %v", sig)
		cli.Unregister()
		os.Exit(0)
	}()

	done := false
	var bgModTime time.Time
	var colors [6][22]int

	// Loop forever; check for a new background every second.
	// When there's a new background, load the image and extract colors, then set keyboard colors to match.
	for {
		for retriesLeft := 2; retriesLeft >= 0; retriesLeft-- {
			m, err := background.ModTime()
			if err == nil {
				if m.After(bgModTime) {
					colors, err = background.Colors()
					if err == nil {
						log.Printf("new background: %v", m)
						bgModTime = m
						break
					}
				}
			}

			// The background image file may be temporarily locked by a Windows process.
			if err != nil {
				log.Printf("error: %v", err)
				log.Printf("retries left: %d", retriesLeft)
				if retriesLeft > 0 {
					time.Sleep(time.Second / 4)
				} else {
					done = true
				}
			}
		}

		if done {
			break
		}

		cli.SetCustom(colors)
		time.Sleep(time.Second)
	}
}
