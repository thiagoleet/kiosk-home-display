package printer

import "log"

// changeLogger reports a recurring failure once, and again only when the
// failure changes. Polling turns any lasting problem — a stopped cupsd, a
// missing lpq — into one identical journal line per interval otherwise.
//
// It holds no lock: every changeLogger belongs to the poll path, which is
// single threaded. Monitor.Start polls before it starts the poll goroutine, so
// the first poll and the ones that follow never overlap.
type changeLogger struct {
	last string
}

// Failed logs the error unless the previous report said the same thing. The
// format is expected to carry a single %v for the error.
func (l *changeLogger) Failed(format string, err error) {
	message := err.Error()

	if message == l.last {
		return
	}

	l.last = message

	log.Printf(format, err)
}

// Recovered logs that the failure is over, and only if there was one.
func (l *changeLogger) Recovered(message string) {
	if l.last == "" {
		return
	}

	l.last = ""

	log.Println(message)
}
