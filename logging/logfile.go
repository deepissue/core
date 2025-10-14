/*
 * Copyright 2022 The Go Authors<36625090@qq.com>. All rights reserved.
 * Use of this source code is governed by a MIT-style
 * license that can be found in the LICENSE file.
 */

package logging

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// LogFile is used to setup a file based logger that also performs log rotation
type LogFile struct {
	//Name of the log file
	name string

	//rotationPolicy
	rotationPolicy rotationPolicy

	//File is the pointer to the current file being written to
	File *os.File

	//Acquire is the mutex utilized to ensure we have no concurrency issues
	acquire sync.Mutex
	fileExt string
}

func (l *LogFile) openNew() error {

	filePointer, err := os.OpenFile(l.name+l.fileExt, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640)
	if err != nil {
		return err
	}

	l.File = filePointer
	return nil
}

// Write is used to implement io.Writer
func (l *LogFile) Write(b []byte) (n int, err error) {
	l.acquire.Lock()
	defer l.acquire.Unlock()

	return l.File.Write(b)
}

func (l *LogFile) Close() {
	if l.File == nil {
		return
	}

	l.File.Close()
}

func (l *LogFile) RotateRename() error {
	l.acquire.Lock()
	defer l.acquire.Unlock()

	//time.RFC3339 2006-01-02T15:04:05Z07:00
	l.File.Sync()
	l.File.Close()
	var past string
	if l.rotationPolicy == rotationPolicyDay {
		past = time.Now().Add(time.Minute * -15).Format("2006-01-02")
	} else {
		past = time.Now().Add(time.Minute * -15).Format("2006-01-02-15")
	}

	newName := l.name + "_" + past + l.fileExt
	_, err := os.Stat(newName)
	if nil == err {
		newName = fmt.Sprintf("%s.%d", newName, time.Now().Unix())
	}
	if err := os.Rename(l.name+l.fileExt, newName); err != nil {
		return err
	}

	return l.openNew()
}
