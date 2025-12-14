// Package plog provides a simple logging system that can buffer messages
// for display in the UI. On Android, messages are also sent to logcat.
package plog

import (
	"fmt"
	"log"
	"time"
)

var (
	pendingMsg = make(chan string, 100)
)

// Printf logs a formatted message and queues it for UI display.
// On Android, this also writes to logcat via the standard log package.
func Printf(format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	// log.Print writes to stderr which Android routes to logcat
	log.Print("mediabackup: " + str)
	msg := fmt.Sprintf("[%s] %s\n", time.Now().Format(time.RFC3339), str)

	select {
	case pendingMsg <- msg:
	default:
	}
}

// MsgChan returns the channel for receiving log messages.
func MsgChan() chan string {
	return pendingMsg
}
