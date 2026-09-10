package printer

// Queue is everything a print system holds at one moment.
//
// Active lists the jobs still to be worked through, in the order the printer
// takes them, so the first entry is the one being printed. Completed lists the
// jobs that have recently left, in no particular order.
//
// Both halves are needed because polling cannot see a short job any other way:
// a job accepted and finished between two polls never appears in Active, and
// only the history it leaves behind proves it ran at all.
type Queue struct {
	Active    []PrintJob
	Completed []PrintJob
}

// Source reports what a print system currently holds. Implementations are
// polled, so Queue must be cheap.
type Source interface {
	Queue() (Queue, error)
}
