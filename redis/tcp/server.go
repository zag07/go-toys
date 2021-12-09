package tcp

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Handler interface {
	Handle(ctx context.Context, conn net.Conn)
	Close() error
}

type Config struct {
	Address    string        `yaml:"address"`
	MaxConnect uint32        `yaml:"max-connect"`
	Timeout    time.Duration `yaml:"timeout"`
}

func ListenAndServe(listener net.Listener, handler Handler, closeChan <-chan struct{}) {
	go func() {
		<-closeChan
		log.Println("closing tcp listener")
		_ = listener.Close()
		_ = handler.Close()
	}()

	defer func() {
		_ = listener.Close()
		_ = handler.Close()
	}()

	ctx := context.Background()
	var waitDone sync.WaitGroup

	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		log.Println("accepted connection")
		waitDone.Add(1)
		go func() {
			defer func() {
				waitDone.Done()
			}()
			handler.Handle(ctx, conn)
		}()
	}
	waitDone.Done()
}

func ListenAndServeWithSignal(cfg *Config, handler Handler) error {
	closeChan := make(chan struct{})
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		sig := <-sigCh
		switch sig {
		case syscall.SIGURG, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			closeChan <- struct{}{}
		}
	}()
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return err
	}
	log.Println("listening on", cfg.Address)
	ListenAndServe(listener, handler, closeChan)
	return nil
}
