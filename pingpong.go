package main

import (
	"fmt"
	"time"
)

func process(kind string, pingChan chan string, pongChan chan string) {
	for {
		if kind == "ping" {
			<-pingChan
			fmt.Println("pong -> ping")
			pongChan <- "pong"
		} else {
			<-pongChan
			fmt.Println("ping -> pong")
			pingChan <- "ping"
		}
		time.Sleep(time.Second)
	}
}

func initPingPong() {
	pingCh := make(chan string, 1)
	pongCh := make(chan string, 1)
	go process("pong", pingCh, pongCh)
	pongCh <- "pong" // seed
	process("ping", pingCh, pongCh)
}
