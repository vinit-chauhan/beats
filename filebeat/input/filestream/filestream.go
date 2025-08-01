// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package filestream

import (
	"compress/gzip"
	"context"
	"errors"
	"io"
	"os"
	"sync/atomic"
	"time"

	"github.com/elastic/go-concert/ctxtool"
	"github.com/elastic/go-concert/timed"
	"github.com/elastic/go-concert/unison"

	input "github.com/elastic/beats/v7/filebeat/input/v2"
	"github.com/elastic/beats/v7/libbeat/common/backoff"
	"github.com/elastic/beats/v7/libbeat/common/file"
	"github.com/elastic/elastic-agent-libs/logp"
)

var (
	ErrClosed       = errors.New("reader closed")
	ErrFileTruncate = errors.New("detected file being truncated")
	ErrInactive     = errors.New("inactive file, reader closed")
)

// logFile contains all log related data
type logFile struct {
	file      File
	log       *logp.Logger
	readerCtx ctxtool.CancelContext

	closeAfterInterval time.Duration
	closeOnEOF         bool

	checkInterval time.Duration
	closeInactive time.Duration
	closeRemoved  bool
	closeRenamed  bool

	isInactive atomic.Bool

	offset       int64
	lastTimeRead time.Time
	backoff      backoff.Backoff
	tg           *unison.TaskGroup
}

// newFileReader creates a new log instance to read log sources
func newFileReader(
	log *logp.Logger,
	canceler input.Canceler,
	f File,
	config readerConfig,
	closerConfig closerConfig,
) (*logFile, error) {
	offset, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	readerCtx := ctxtool.WithCancelContext(ctxtool.FromCanceller(canceler))
	tg := unison.TaskGroupWithCancel(readerCtx)

	l := &logFile{
		file:               f,
		log:                log,
		closeAfterInterval: closerConfig.Reader.AfterInterval,
		closeOnEOF:         closerConfig.Reader.OnEOF,
		checkInterval:      closerConfig.OnStateChange.CheckInterval,
		closeInactive:      closerConfig.OnStateChange.Inactive,
		closeRemoved:       closerConfig.OnStateChange.Removed,
		closeRenamed:       closerConfig.OnStateChange.Renamed,
		offset:             offset,
		lastTimeRead:       time.Now(),
		backoff:            backoff.NewExpBackoff(canceler.Done(), config.Backoff.Init, config.Backoff.Max),
		readerCtx:          readerCtx,
		tg:                 tg,
	}

	l.startFileMonitoringIfNeeded()

	return l, nil
}

// Read reads from the reader and updates the offset
// The total number of bytes read is returned.
// If the file is inactive, ErrInactive is returned
// If f.readerCtx is cancelled for any reason, then
// ErrClosed is returned
func (f *logFile) Read(buf []byte) (int, error) {
	totalN := 0

	for f.readerCtx.Err() == nil {
		n, err := f.file.Read(buf)
		if n > 0 {
			f.offset += int64(n)
			f.lastTimeRead = time.Now()
		}
		totalN += n

		// Read from source completed without error
		// Either end reached or buffer full
		if err == nil {
			// reset backoff for next read
			f.backoff.Reset()
			return totalN, nil
		}

		// Move buffer forward for next read
		buf = buf[n:]

		// Checks if an error happened or buffer is full
		// If buffer is full, cannot continue reading.
		// Can happen if n == bufferSize + io.EOF error
		err = f.errorChecks(err)
		if err != nil || len(buf) == 0 {
			return totalN, err
		}

		f.log.Debugf("End of file reached: %s; Backoff now.", f.file.Name())
		f.backoff.Wait()
	}

	// `f.isInactive` is set in a different goroutine, however it is the same
	// goroutine that cancels `f.readerCtx`. The timer goroutine that checks
	// for inactivity first sets `f.isInactive`, then it cancels `f.readerCtx`,
	// once the execution comes back to this method, the for loop condition
	// evaluates to false and the correct value from `f.isInactive` is read.
	if f.isInactive.Load() {
		return 0, ErrInactive
	}

	return 0, ErrClosed
}

func (f *logFile) startFileMonitoringIfNeeded() {
	if f.closeInactive > 0 || f.closeRemoved || f.closeRenamed {
		err := f.tg.Go(func(ctx context.Context) error {
			f.periodicStateCheck(ctx)
			return nil
		})
		if err != nil {
			f.log.Errorf("failed to start file monitoring: %w", err)
		}
	}

	if f.closeAfterInterval > 0 {
		err := f.tg.Go(func(ctx context.Context) error {
			f.closeIfTimeout(ctx)
			return nil
		})
		if err != nil {
			f.log.Errorf("failed to schedule a file close: %w", err)
		}
	}
}

func (f *logFile) closeIfTimeout(ctx unison.Canceler) {
	if err := timed.Wait(ctx, f.closeAfterInterval); err == nil {
		f.readerCtx.Cancel()
	}
}

func (f *logFile) periodicStateCheck(ctx unison.Canceler) {
	err := timed.Periodic(ctx, f.checkInterval, func() error {
		if f.shouldBeClosed() {
			f.readerCtx.Cancel()
		}
		return nil
	})
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			f.log.Errorf("failed to schedule a periodic state check: %s", err)
		}
	}
}

func (f *logFile) shouldBeClosed() bool {
	if f.closeInactive > 0 {
		if time.Since(f.lastTimeRead) > f.closeInactive {
			f.isInactive.Store(true)
			f.log.Debugf("'%s' is inactive", f.file.Name())
			return true
		}
	}

	if !f.closeRemoved && !f.closeRenamed {
		return false
	}

	info, statErr := f.file.Stat()
	if statErr != nil {
		// return early if the file does not exist anymore and the reader should be closed
		if f.closeRemoved && errors.Is(statErr, os.ErrNotExist) {
			f.log.Debugf("close.on_state_change.removed is enabled and file %s has been removed", f.file.Name())
			return true
		}

		// If an unexpected error happens we keep the reader open hoping once everything will go back to normal.
		f.log.Errorf("Unexpected error reading from %s; error: %s", f.file.Name(), statErr)
		return false
	}

	if f.closeRenamed {
		// Check if the file can still be found under the same path
		if !isSameFile(f.file.Name(), info) {
			f.log.Debugf("close.on_state_change.renamed is enabled and file %s has been renamed", f.file.Name())
			return true
		}
	}

	if f.closeRemoved {
		// Check if the file name exists. See https://github.com/elastic/filebeat/issues/93
		if file.IsRemoved(f.file.OSFile()) {
			f.log.Debugf("close.on_state_change.removed is enabled and file %s has been removed", f.file.Name())
			return true
		}
	}

	return false
}

func isSameFile(path string, info os.FileInfo) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}

	return os.SameFile(fileInfo, info)
}

// errorChecks determines the cause for EOF errors, and how the EOF event should be handled
// based on the config options.
func (f *logFile) errorChecks(err error) error {
	if !errors.Is(err, io.EOF) {
		f.log.Errorf("Unexpected state reading from %s; error: %s",
			f.file.Name(), err)

		// gzip.ErrChecksum happens after all data is read from a GZIP file, and
		// it's recoverable, nothing else to do. Thus, we return EOF.
		if errors.Is(err, gzip.ErrChecksum) {
			return io.EOF
		}

		return err
	}

	return f.handleEOF()
}

func (f *logFile) handleEOF() error {
	if f.closeOnEOF || f.file.IsGZIP() {
		return io.EOF
	}

	// Re-fetch fileinfo to check if the file was truncated.
	// Errors if the file was removed/rotated after reading and before
	// calling the stat function
	info, statErr := f.file.Stat()
	if statErr != nil {
		f.log.Error("Unexpected error reading from %s; error: %s", f.file.Name(), statErr)
		return statErr
	}

	if info.Size() < f.offset {
		f.log.Debugf("File was truncated as offset (%d) > size (%d): %s", f.offset, info.Size(), f.file.Name())
		return ErrFileTruncate
	}

	return nil
}

// Close
func (f *logFile) Close() error {
	f.readerCtx.Cancel()
	err := f.file.Close()
	_ = f.tg.Stop() // Wait until all resources are released for sure.
	f.log.Debugf("Closed reader. Path='%s'", f.file.Name())
	return err
}
