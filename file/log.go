package file

import (
	"github.com/fsnotify/fsnotify"
	"log"
)

// NotifyLog provides a way for filesystem changes to be logged.
// This allows for easily plugging in your logger of choice.
type NotifyLog interface {
	Event(event fsnotify.Event)
	UnrelatedEvent(event fsnotify.Event)
	Error(err error)
}

// StdLog creates an instance of NotifyLog that uses the standard [log] package.
// Unrelated events will not be logged.
func StdLog() NotifyLog {
	return stdLog{}
}

var _ NotifyLog = (*stdLog)(nil)

type stdLog struct{}

func (s stdLog) Event(event fsnotify.Event) {
	log.Printf("fs-event: [%s] %s\n", event.Op.String(), event.Name)
}

func (s stdLog) UnrelatedEvent(_ fsnotify.Event) {
	// Nothing logged for unrelated events.
}

func (s stdLog) Error(err error) {
	log.Printf("fs-event: ERROR: %v\n", err)
}

type noOpNofifyLog struct{}

func (n *noOpNofifyLog) Event(_ fsnotify.Event) {}

func (n *noOpNofifyLog) UnrelatedEvent(_ fsnotify.Event) {}

func (n *noOpNofifyLog) Error(_ error) {}

func newNoOpNotifyLog() NotifyLog {
	return &noOpNofifyLog{}
}
