/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"zstdb/cmd"
)

func main() {
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		cmd.Execute()
	}()

	go func() {
		cmd.BadgerRunValueLogGC()
	}()

	go func() {
		onExit()
	}()

	wg.Wait()
}

func onExit() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cmd.StopGrpcServer()
		time.Sleep(time.Second * 2)
		os.Exit(0)
	}()
}
