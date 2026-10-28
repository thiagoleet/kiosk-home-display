package news

import (
	"fmt"
	"strings"
	"time"
)

const DefaultCategory = "general"

// Source is a configured feed. ID is the database row, assigned when the
// configured sources are synced at startup.
type Source struct {
	ID            int64
	Name          string
	FeedURL       string
	Category      string
	Enabled       bool
	LastFetchedAt *time.Time
	LastError     string
}

// ParseSources reads the NEWS_SOURCES format: entries separated by ";", each
// one "name|feed url" or "name|feed url|category".
//
//	G1|https://g1.globo.com/rss/g1/|general;NASA|https://www.nasa.gov/feed/|science
func ParseSources(value string) ([]Source, error) {
	var sources []Source

	seen := map[string]bool{}

	for entry := range strings.SplitSeq(value, ";") {
		entry = strings.TrimSpace(entry)

		if entry == "" {
			continue
		}

		fields := strings.Split(entry, "|")

		if len(fields) < 2 || len(fields) > 3 {
			return nil, fmt.Errorf(
				"invalid news source %q, want name|url or name|url|category",
				entry,
			)
		}

		source := Source{
			Name:     strings.TrimSpace(fields[0]),
			FeedURL:  strings.TrimSpace(fields[1]),
			Category: DefaultCategory,
			Enabled:  true,
		}

		if len(fields) == 3 {
			if category := strings.TrimSpace(fields[2]); category != "" {
				source.Category = category
			}
		}

		if source.Name == "" {
			return nil, fmt.Errorf(
				"news source %q has no name",
				entry,
			)
		}

		if httpURL(source.FeedURL) == "" {
			return nil, fmt.Errorf(
				"news source %q has an invalid feed url",
				source.Name,
			)
		}

		if seen[source.FeedURL] {
			return nil, fmt.Errorf(
				"news feed %q is listed twice",
				source.FeedURL,
			)
		}

		seen[source.FeedURL] = true

		sources = append(sources, source)
	}

	return sources, nil
}
