package agent

import (
	"compress/gzip"
	"fmt"
	"io"
	"net"
)

// type bufferedGZIPReader struct {
// 	bufioReader *bufio.Reader
// 	gzipReader  *gzip.Reader
// }

// func newBufferedGZIPReader(r io.Reader) *bufferedGZIPReader {
// 	return &bufferedGZIPReader{
// 		bufio: bufio.NewReaderSize(r, 10),
// 	}
// }

// func (r bufferedGZIPReader) Read(b []byte) (int, error) {
// 	if r.gzipReader != nil {
// 		return r.gzipReader.Read(b)
// 	}
// 	b, err := r.bufioReader.Peek(10)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if len(b) == 10 {

// 	}
// }

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
	fmt.Println("FLUSH readWriter outside")
	if flusher, ok := rw.Writer.(interface{ Flush() error }); ok {
		fmt.Println("FLUSH readWriter inside")
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
	fmt.Println("FLUSH agent outside")
	if flusher, ok := a.conn.(interface{ Flush() error }); ok {
		fmt.Println("FLUSH agent inside")
		err := flusher.Flush()
		if err != nil {
			panic(err)
		}
	}
}
