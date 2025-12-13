// Package network provides network state abstractions that can be set
// from the Android layer.
package network

import (
	"fmt"
	"sync"
)

// ConnState represents the network connection state.
type ConnState int

func (cs ConnState) String() string {
	switch cs {
	case ConnStateUnknown:
		return "ConnStateUnknown"
	case NoNetwork:
		return "NoNetwork"
	case NotWifi:
		return "NotWifi"
	case Wifi:
		return "Wifi"
	default:
		return fmt.Sprintf("ConnStateUnknown<%d>", cs)
	}
}

const (
	ConnStateUnknown ConnState = -2
	NoNetwork        ConnState = -1
	NotWifi          ConnState = 0
	Wifi             ConnState = 1
)

var (
	currentState ConnState = ConnStateUnknown
	stateMu      sync.RWMutex
)

// SetConnectionState sets the current network connection state.
// This should be called from the Android layer when network state changes.
func SetConnectionState(state ConnState) {
	stateMu.Lock()
	defer stateMu.Unlock()
	currentState = state
}

// GetConnectionState returns the current network connection state.
func GetConnectionState() ConnState {
	stateMu.RLock()
	defer stateMu.RUnlock()
	return currentState
}
