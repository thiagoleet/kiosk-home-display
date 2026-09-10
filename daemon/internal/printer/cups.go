package printer

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// queuedJob matches the identifier lpstat prints in the first column of a
// job, which is the destination and the CUPS job number joined by a dash,
// such as "Brother_DCP_1600_series-31". Anything else on the line is a header
// or a warning and is skipped.
var queuedJob = regexp.MustCompile(
	`^(.+)-([0-9]+)$`,
)

// CUPSSource reads the queue through the CUPS command line tools, the same way
// the display controllers drive xset and wlopm. Nothing links against libcups,
// so the daemon runs against whatever cupsd the host already has.
//
// lpstat is the source of truth: it prints one line per job and its first
// column is a stable, unique identifier. It does not print document titles
// though, so lpq fills those in, and a job whose title cannot be recovered
// keeps its identifier as a name.
//
// The two loggers are only touched from the poll path, which is single
// threaded. See changeLogger.
type CUPSSource struct {
	command commandRunner

	titleLog     changeLogger
	completedLog changeLogger
}

func NewCUPSSource() *CUPSSource {
	return &CUPSSource{
		command: runCommand,
	}
}

func (s *CUPSSource) Queue() (Queue, error) {
	active, err := s.jobs("not-completed")
	if err != nil {
		return Queue{}, err
	}

	// lpq lists the queue, so only a job still in it can be given a title.
	if len(active) > 0 {
		s.applyTitles(active)
	}

	// The history is what catches a job that came and went inside one poll
	// interval. Losing it costs those short jobs, which is not worth failing
	// the whole poll over while the active queue still reads correctly.
	completed, err := s.jobs("completed")
	if err != nil {
		s.completedLog.Failed(
			"[PRINTER] finished jobs unavailable, prints shorter than one poll will be missed: %v",
			err,
		)

		return Queue{Active: active}, nil
	}

	s.completedLog.Recovered(
		"[PRINTER] finished jobs readable again",
	)

	return Queue{
		Active:    active,
		Completed: completed,
	}, nil
}

// jobs reads one half of the queue. The selector is the lpstat -W argument:
// "not-completed" for the jobs still to run, "completed" for the history.
func (s *CUPSSource) jobs(
	selector string,
) ([]PrintJob, error) {
	output, err := s.command(
		"lpstat",
		"-W",
		selector,
		"-o",
	)
	if err != nil {
		return nil, cupsError("lpstat", err)
	}

	return parseQueuedJobs(output), nil
}

// applyTitles replaces the identifiers standing in as names with the document
// titles, for the jobs lpq can still see. It is best effort: losing a title is
// not worth failing a poll over, so a broken lpq leaves every job named after
// its identifier.
func (s *CUPSSource) applyTitles(jobs []PrintJob) {
	output, err := s.command("lpq", "-a")
	if err != nil {
		s.titleLog.Failed(
			"[PRINTER] job titles unavailable: %v",
			err,
		)

		return
	}

	s.titleLog.Recovered(
		"[PRINTER] job titles readable again",
	)

	titles := parseJobTitles(output)

	for index, job := range jobs {
		number := strconv.Itoa(job.Number)

		if title := titles[number]; title != "" {
			jobs[index].Name = title
		}
	}
}

// parseQueuedJobs reads the job listing lpstat prints, which is the same for
// the queue and for the history:
//
//	Brother_DCP_1600_series-31 thiago 485376 Wed Sep  9 15:43:15 2026
//
// Only the identifier is taken from it, and the job number within it. The
// remaining columns describe the owner, the byte size and the queue time,
// none of which the display shows.
func parseQueuedJobs(output string) []PrintJob {
	var jobs []PrintJob

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)

		if len(fields) < 3 {
			continue
		}

		matches := queuedJob.FindStringSubmatch(fields[0])

		if matches == nil {
			continue
		}

		number, err := strconv.Atoi(matches[2])
		if err != nil {
			continue
		}

		jobs = append(jobs, PrintJob{
			ID:     fields[0],
			Name:   fields[0],
			Number: number,
		})
	}

	return jobs
}

// parseJobTitles reads the queue listing lpq prints:
//
//	Rank    Owner   Job     File(s)                         Total Size
//	active  thiago  31      report.pdf                      485376 bytes
//
// The title sits between fixed columns and may itself contain spaces, so it is
// whatever is left after the leading rank, owner and job number and the
// trailing size. Titles are truncated by lpq at 31 characters.
func parseJobTitles(output string) map[string]string {
	titles := make(map[string]string)

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)

		if len(fields) < 5 {
			continue
		}

		if fields[len(fields)-1] != "bytes" {
			continue
		}

		number := fields[2]

		if !isNumber(number) {
			continue
		}

		title := strings.Join(
			fields[3:len(fields)-2],
			" ",
		)

		if title == "" {
			continue
		}

		titles[number] = title
	}

	return titles
}

func isNumber(value string) bool {
	if value == "" {
		return false
	}

	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}

	return true
}

// cupsError names the way CUPS monitoring fails on a fresh box. A missing
// package looks like a plain exit status in the journal otherwise, which is the
// hardest part of the problem to spot.
func cupsError(name string, err error) error {
	if errors.Is(err, exec.ErrNotFound) {
		return fmt.Errorf(
			"%s is not installed, install cups-client: %w",
			name,
			err,
		)
	}

	return err
}
