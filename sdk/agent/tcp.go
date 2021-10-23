package agent

import (
	"compress/gzip"
	"fmt"
	"io"
	"net"
)

type readWriter struct {
	io.Reader
	io.Writer
}

func (a *Agent) ServeTCP(addr string) error {
	if a.conn != nil {
		return fmt.Errorf("already connected")
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listening on %s: %w", addr, err)
	}
	conn, err := ln.Accept()
	if err != nil {
		return fmt.Errorf("accepting incoming connection: %w", err)
	}
	fmt.Fprintf(a.logWriter, "accepted connection from %v\n", conn.RemoteAddr())

	zw, err := gzip.NewWriterLevel(conn, gzip.BestSpeed)
	if err != nil {
		return fmt.Errorf("creating gzip writer: %w", err)
	}
	zr, err := gzip.NewReader(conn)
	if err != nil {
		return fmt.Errorf("creating gzip reader: %w", err)
	}
	a.conn = readWriter{
		Reader: zr,
		Writer: zw,
	}

	err = a.hello()
	if err != nil {
		return fmt.Errorf("sending hello: %w", err)
	}
	go a.receiveLoop()
	return nil
}

func (a *Agent) ConnectTCP(addr string) error {
	if a.conn != nil {
		return fmt.Errorf("already connected")
	}
	var err error
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("connecting to %s: %w", addr, err)
	}
	fmt.Fprintf(a.logWriter, "connected to %v\n", conn.RemoteAddr())
	a.conn = conn
	err = a.hello()
	if err != nil {
		return fmt.Errorf("sending hello: %w", err)
	}
	go a.receiveLoop()
	return nil
}

func (a *Agent) connFlush() {
	if flusher, ok := a.conn.(interface{ Flush() error }); ok {
		err := flusher.Flush()
		if err != nil {
			panic(err)
		}
	}
}
