package printer

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
)

// queuedJob matches the identifier lpstat prints in the first column of a
// queued job, which is the destination and the CUPS job number joined by a
// dash, such as "Brother_DCP_1600_series-31". Anything else on the line is a
// header or a warning and is skipped.
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
type CUPSSource struct {
	command commandRunner
}

func NewCUPSSource() *CUPSSource {
	return &CUPSSource{
		command: runCommand,
	}
}

func (s *CUPSSource) ActiveJobs() ([]PrintJob, error) {
	output, err := s.command(
		"lpstat",
		"-W",
		"not-completed",
		"-o",
	)
	if err != nil {
		return nil, cupsError("lpstat", err)
	}

	jobs := parseQueuedJobs(output)

	if len(jobs) == 0 {
		return nil, nil
	}

	titles := s.jobTitles()

	for index, job := range jobs {
		number := jobNumber(job.ID)

		if title := titles[number]; title != "" {
			jobs[index].Name = title
		}
	}

	return jobs, nil
}

// jobTitles maps CUPS job numbers to document titles. It is best effort: lpq
// is the only tool that reports titles, and losing them is not worth failing a
// poll over, so a broken lpq leaves every job named after its identifier.
func (s *CUPSSource) jobTitles() map[string]string {
	output, err := s.command("lpq", "-a")
	if err != nil {
		log.Printf(
			"[PRINTER] job titles unavailable: %v",
			err,
		)

		return nil
	}

	return parseJobTitles(output)
}

// parseQueuedJobs reads the job listing lpstat prints:
//
//	Brother_DCP_1600_series-31 thiago 485376 Wed Sep  9 15:43:15 2026
//
// Only the identifier is taken from it. The remaining columns describe the
// owner, the byte size and the queue time, none of which the display shows.
func parseQueuedJobs(output string) []PrintJob {
	var jobs []PrintJob

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)

		if len(fields) < 3 {
			continue
		}

		if !queuedJob.MatchString(fields[0]) {
			continue
		}

		jobs = append(jobs, PrintJob{
			ID:   fields[0],
			Name: fields[0],
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

// jobNumber returns the CUPS job number carried by an lpstat identifier, which
// is what lpq reports jobs by.
func jobNumber(id string) string {
	matches := queuedJob.FindStringSubmatch(id)

	if matches == nil {
		return ""
	}

	return matches[2]
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
