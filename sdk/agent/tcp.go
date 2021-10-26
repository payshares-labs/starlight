package agent

import (
	"compress/gzip"
	"fmt"
	"io"
	"net"
)

type lazyReader struct {
	makeReader func() (io.Reader, error)
	reader     io.Reader
}

func newLazyReader(makeReader func() (io.Reader, error)) *lazyReader {
	return &lazyReader{
		makeReader: makeReader,
	}
}

func (r lazyReader) Read(b []byte) (int, error) {
	if r.reader == nil {
		reader, err := r.makeReader()
		if err != nil {
			return 0, err
		}
		r.reader = reader
	}
	return r.reader.Read(b)
}

type readWriter struct {
	io.Reader
	io.Writer
}

func (rw readWriter) Flush() error {
	if flusher, ok := rw.Writer.(interface{ Flush() error }); ok {
		err := flusher.Flush()
		if err != nil {
			return err
		}
	}
	return nil
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
	r := newLazyReader(func() (io.Reader, error) {
		return gzip.NewReader(conn)
	})
	a.conn = readWriter{
		Reader: r,
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

	zw, err := gzip.NewWriterLevel(conn, gzip.BestSpeed)
	if err != nil {
		return fmt.Errorf("creating gzip writer: %w", err)
	}
	r := newLazyReader(func() (io.Reader, error) {
		return gzip.NewReader(conn)
	})
	a.conn = readWriter{
		Reader: r,
		Writer: zw,
	}

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
