package printer

// Source lists the jobs a print system currently holds. Implementations are
// polled, so ActiveJobs must be cheap and must return the queue in the order
// the jobs are worked through.
type Source interface {
	ActiveJobs() ([]PrintJob, error)
}
