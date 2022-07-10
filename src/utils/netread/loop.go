package netread

import (
	"log"
	"net"
	"time"
)

// ReadLoop is used to read from a net.Conn in a select statement. Data is
// read into a single buffer to avoid large amounts of allocations. Becuase of
// this, reading must happen somewhat synchronously; first check for available
// data using ReadChBuf, then call Ready once data has been processed and
// is ready to be overwritten with new data.
type ReadLoop struct {
	buf     []byte        // Data read from conn
	chBuf   chan []byte   // Written to when conn.Read has data
	chReady chan struct{} // Once data in chBuf is processed, this signals user is ready to read again
	chErr   chan error    // Errors encountered when reading
	chDone  chan struct{} // Once user is done, Run() will exit

	conn    net.Conn
	timeout time.Duration
}

func NewReadLoop(bufSize int64, conn net.Conn, timeout time.Duration) *ReadLoop {
	return &ReadLoop{
		buf:     make([]byte, bufSize, bufSize),
		chBuf:   make(chan []byte, 1),
		chReady: make(chan struct{}, 1),
		chErr:   make(chan error),
		chDone:  make(chan struct{}),
		conn:    conn,
		timeout: timeout,
	}
}

// ReadChBuf returns the read end of the buffer channel.
func (rl *ReadLoop) ReadChBuf() <-chan []byte {
	return rl.chBuf
}

// Ready tells the ReadLoop that the data in the buffer channel has been
// processed and that we are ready to read more data from conn.
func (rl *ReadLoop) Ready() {
	rl.chReady <- struct{}{}
}

// ReadChErr returns the read end of the error channel.
func (rl *ReadLoop) ReadChErr() <-chan error {
	return rl.chErr
}

// Finish cancels the current conn.Read() call and terminates the call to Run().
func (rl *ReadLoop) Finish() {
	rl.chDone <- struct{}{}
	// Set the cancel time to now. Rather than using time.Now() which involves
	// a syscall, use Unix(1,0)
	_ = rl.conn.SetReadDeadline(time.Unix(1, 0))
}

// Run will enter an infinite loop reading from the connection. When data is
// read, it will write a byte slice to the buffer channel, which is accessed
// via ReadChBuf. If there are any errors, they are reported on the error
// channel, accessed via ReadChErr. To finish the loop, call Finish.
func (rl *ReadLoop) Run() {
	log.Printf("started read loop for %v", rl.conn.RemoteAddr())
	defer log.Printf("ended read loop for %v", rl.conn.RemoteAddr())

	done := false
	for !done {
		e := rl.conn.SetReadDeadline(time.Now().Add(rl.timeout))
		if e != nil {
			rl.chErr <- e
			break
		}

		n, e := rl.conn.Read(rl.buf)
		if n == 0 || e != nil {
			rl.chErr <- e
			break
		} else {
			rl.chBuf <- rl.buf[:n]

			select {
			case <-rl.chReady:
			case <-rl.chDone:
				done = true
			}
		}
	}
}
