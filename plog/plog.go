// Package plog provides a simple logging system that can buffer messages
// for display in the UI.
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
func Printf(format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	log.Print(str)
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
