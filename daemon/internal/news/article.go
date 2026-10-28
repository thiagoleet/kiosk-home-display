package news

import (
	"crypto/sha1"
	"encoding/hex"
	"net/url"
	"strings"
	"time"
)

// Article is the shape every RSS and Atom entry is normalised into, whatever
// the feed it came from.
type Article struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary,omitempty"`
	Source      string    `json:"source"`
	Category    string    `json:"category,omitempty"`
	ImageURL    string    `json:"imageUrl,omitempty"`
	URL         string    `json:"url"`
	PublishedAt time.Time `json:"publishedAt"`

	// CanonicalURL is the normalised link. Two sources carrying the same story
	// store a row each, and the listing shows only one of them.
	CanonicalURL string `json:"-"`
}

// Feed is what the slideshow reads, taken from a single read transaction so
// the articles, the update time and the source statuses agree with each other.
//
// UpdatedAt is when the last refresh cycle in which at least one source was
// fetched successfully completed. It is nil until that has ever happened, so
// the frontend can tell stale content apart from a fresh poll of the endpoint.
type Feed struct {
	Articles  []Article      `json:"articles"`
	UpdatedAt *time.Time     `json:"updatedAt"`
	Sources   []SourceStatus `json:"sources"`
}

// Source statuses as the frontend sees them. The error text itself stays in
// the database and the journal: it names hosts and resolver details the kiosk
// has no use for.
const (
	SourceStatusOK      = "ok"
	SourceStatusError   = "error"
	SourceStatusPending = "pending"
)

type SourceStatus struct {
	Name          string     `json:"name"`
	Category      string     `json:"category"`
	Status        string     `json:"status"`
	LastSuccessAt *time.Time `json:"lastSuccessAt"`
}

// articleID derives an identifier that survives successive fetches of the same
// feed, so a story that is still listed an hour later updates its row instead
// of appearing twice.
//
// Both keys are scoped to the feed. GUIDs need it because many feeds number
// their items with short values that would collide across sources. Links need
// it so that a story two sources share is owned by each of them: disabling one
// source must not take the other's copy with it.
func articleID(
	feedURL string,
	guid string,
	link string,
) string {
	key := ""

	switch guid = strings.TrimSpace(guid); {
	case guid != "":
		key = "guid:" + feedURL + "\n" + guid

	case link != "":
		key = "url:" + feedURL + "\n" + normalizeURL(link)

	default:
		return ""
	}

	sum := sha1.Sum([]byte(key))

	return hex.EncodeToString(sum[:])
}

// normalizeURL removes the parts of a link that change between fetches without
// changing the article: the fragment, tracking parameters and letter case of
// the scheme and host.
func normalizeURL(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return strings.TrimSpace(raw)
	}

	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Fragment = ""

	query := parsed.Query()

	for key := range query {
		lower := strings.ToLower(key)

		if strings.HasPrefix(lower, "utm_") ||
			lower == "fbclid" ||
			lower == "gclid" {
			query.Del(key)
		}
	}

	parsed.RawQuery = query.Encode()
	parsed.Path = strings.TrimSuffix(parsed.Path, "/")

	return parsed.String()
}

// httpURL returns the link only when it is an absolute http or https URL, so a
// javascript: or relative link from a feed never reaches the kiosk browser.
func httpURL(raw string) string {
	raw = strings.TrimSpace(raw)

	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ""
	}

	if parsed.Host == "" {
		return ""
	}

	return raw
}

// resolveURL turns a link from a feed into an absolute http or https URL.
// Relative and protocol-relative references are resolved against base, which
// is the feed's own URL, an xml:base, or the article page; without a base, or
// when the result is any other scheme, the link is dropped.
func resolveURL(base *url.URL, raw string) string {
	raw = strings.TrimSpace(raw)

	if raw == "" {
		return ""
	}

	reference, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	if base != nil {
		reference = base.ResolveReference(reference)
	}

	return httpURL(reference.String())
}

// parseBase parses a resolved URL for use as the base of further references.
func parseBase(raw string) *url.URL {
	if raw == "" {
		return nil
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return nil
	}

	return parsed
}
