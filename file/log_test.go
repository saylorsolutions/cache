package file

import (
	"errors"
	"github.com/fsnotify/fsnotify"
	"testing"
)

var _ NotifyLog = (*testNotifyLog)(nil)

type testNotifyLog struct {
	t           *testing.T
	failOnError bool
}

func testingLog(t *testing.T) NotifyLog {
	return &testNotifyLog{t: t, failOnError: true}
}

func (t *testNotifyLog) Event(event fsnotify.Event) {
	t.t.Logf("Received event [%s] %s", event.Op.String(), event.Name)
}

func (t *testNotifyLog) UnrelatedEvent(event fsnotify.Event) {
	t.t.Logf("Received unrelated event [%s] %s", event.Op.String(), event.Name)
}

func (t *testNotifyLog) Error(err error) {
	var log func(msg string, args ...any)
	if t.failOnError {
		log = t.t.Errorf
	} else {
		log = t.t.Logf
	}
	log("Error: %v", err)
}

func ExampleStdLog() {
	log := StdLog()

	// Called by the fsnotify goroutine when a filesystem event is dispatched.
	log.Event(fsnotify.Event{
		Name: "some-file.txt",
		Op:   fsnotify.Write,
	})

	// Called by the fsnotify goroutine.
	log.Error(errors.New("something bad happened"))

	// Output:
}
