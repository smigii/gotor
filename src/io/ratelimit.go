package io

import (
	"io"
	"sync"
	"sync/atomic"
	"time"
)

const (
	NoLimit = -1
)

type RateLimitIO struct {
	rRate     int64
	rUsed     int64
	rlock     sync.Mutex
	wRate     int64
	wUsed     int64
	wlock     sync.Mutex
	sig       chan struct{} // Signal that controls the run() loop
	tick      *time.Ticker
	rTotal    int64
	wTotal    int64
	timeTotal time.Duration
}

func NewRateLimitIO() *RateLimitIO {
	return &RateLimitIO{
		rRate: NoLimit,
		wRate: NoLimit,
		rUsed: 0,
		wUsed: 0,
		sig:   make(chan struct{}),
		tick:  time.NewTicker(1 * time.Second),
	}
}

func (rl *RateLimitIO) SetReadRate(limit int64) {
	if limit < NoLimit {
		limit = NoLimit
	}
	rl.rlock.Lock()
	defer rl.rlock.Unlock()
	rl.rRate = limit // Only place this is modified, don't need atomic
}

func (rl *RateLimitIO) SetWriteRate(limit int64) {
	if limit < NoLimit {
		limit = NoLimit
	}
	rl.wlock.Lock()
	defer rl.wlock.Unlock()
	rl.wRate = limit // Only place this is modified, don't need atomic
}

// Run will maintain the set rate limit in an infinite loop, until Stop() is
// called. This method should almost certainly be in its own goroutine.
func (rl *RateLimitIO) Run() {
	rl.rUsed = 0
	rl.wUsed = 0

	done := false
	for !done {
		select {
		case <-rl.sig:
			done = true
			break
		case <-rl.tick.C:
			atomic.StoreInt64(&rl.wUsed, 0)
			atomic.StoreInt64(&rl.rUsed, 0)
		}
	}
}

func (rl *RateLimitIO) Stop() {
	rl.sig <- struct{}{}
}

// Write is a blocking call that will write the data to the io.Writer in a
// rate-limitted manner. For now, data that is larger than the write rate
// will be written in full, rather than broken apart. I.e, trying to write
// 20 bytes when the write rate is 5 bytes will write the full 20 bytes,
// rather than splitting into 4 seperate writes.
func (rl *RateLimitIO) Write(writer io.Writer, data []byte) error {
	lenData := int64(len(data))

	// No sense locking here, if we introduce a limit during a call to Write,
	// the lock would only be acquired after the in-progress Write finishes
	// and the rate-limit would be effectively ignored.
	if rl.wRate == NoLimit {
		_, e := writer.Write(data)
		atomic.AddInt64(&rl.wUsed, lenData)
		atomic.AddInt64(&rl.wTotal, lenData)
		return e
	}

	var unlocker sync.Once

	rl.wlock.Lock()
	defer unlocker.Do(rl.wlock.Unlock)

	for {
		if rl.wUsed < rl.wRate {
			break
		}
	}

	atomic.AddInt64(&rl.wUsed, lenData)
	unlocker.Do(rl.wlock.Unlock)
	atomic.AddInt64(&rl.wTotal, lenData)

	_, e := writer.Write(data)
	return e
}

// Read is a blocking call that will read from the io.Reader into the buffer
// in a rate-limitted manner. For now, data that is larger than the read
// rate will be written in full, rather than broken apart. I.e, trying to read
// 20 bytes when the read rate is 5 bytes will read the full 20 bytes,
// rather than splitting into 4 seperate reads.
func (rl *RateLimitIO) Read(reader io.Reader, buf []byte) error {
	lenBuf := int64(len(buf))

	// No sense locking here, if we introduce a limit during a call to Read,
	// the lock would only be acquired after the in-progress Read finishes
	// and the rate-limit would be effectively ignored.
	if rl.rRate == NoLimit {
		_, e := reader.Read(buf)
		atomic.AddInt64(&rl.rUsed, lenBuf)
		atomic.AddInt64(&rl.rTotal, lenBuf)
		return e
	}

	var unlocker sync.Once

	rl.rlock.Lock()
	defer unlocker.Do(rl.rlock.Unlock)

	for {
		if rl.rUsed < rl.rRate {
			break
		}
	}

	atomic.AddInt64(&rl.rUsed, lenBuf)
	unlocker.Do(rl.rlock.Unlock)
	atomic.AddInt64(&rl.rTotal, lenBuf)

	_, e := reader.Read(buf)
	return e
}
