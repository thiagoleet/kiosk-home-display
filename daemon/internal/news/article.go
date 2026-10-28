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
}

// Feed is what the slideshow reads. UpdatedAt is the last successful fetch of
// any enabled source, nil until one has ever succeeded, so the frontend can
// tell stale content apart from a fresh poll of the endpoint.
type Feed struct {
	Articles  []Article  `json:"articles"`
	UpdatedAt *time.Time `json:"updatedAt"`
}

// articleID derives an identifier that survives successive fetches of the same
// feed, so a story that is still listed an hour later updates its row instead
// of appearing twice.
//
// The GUID is preferred, scoped to the feed because many feeds number their
// items with short values that would collide across sources. Without one the
// normalised link stands in, unscoped, since a URL is already unique.
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
		key = "url:" + normalizeURL(link)

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
