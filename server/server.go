package server

import (
	"fmt"
	"net"

	"github.com/joetang09/http-buffer-agent/config"
)

var (
	redirect *Redirect
	listener *net.TCPListener
	stopped  = true
)

// Serve server start
func Serve() (err error) {
	redirect = NewRedirect(config.GetInt("retrytimes"), config.GetInt("outparallel"), config.GetInt("bufferlength"))
	addr, _ := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", config.GetInt("port")))
	listener, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	stopped = false
	if err = redirect.Run(); err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
		}
		if stopped {
			break
		}
		go redirect.Forward(conn)
	}

	return nil
}

// Shutdown server stop
func Shutdown() error {
	stopped = true
	if err := listener.Close(); err != nil {
		return err
	}
	redirect.Stop()
	return nil
}
