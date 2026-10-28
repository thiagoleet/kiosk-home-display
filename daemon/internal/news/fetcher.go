package news

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// maxFeedSize bounds how much of a response is read. Real feeds are a few
// hundred kilobytes; anything far beyond that is a misconfigured URL serving
// something else, and the Pi has no memory to spare for it.
const maxFeedSize = 5 << 20

// maxSummaryLength keeps summaries to what a slide can show. Some feeds put the
// whole article body in the description.
const maxSummaryLength = 280

const userAgent = "KioskHomeDisplay/1.0 (+https://github.com/thiagoleet/kiosk-home-display)"

var ErrNotAFeed = errors.New("response is not an RSS or Atom feed")

type Fetcher interface {
	Fetch(ctx context.Context, source Source) ([]Article, error)
}

// imageLookupTimeout bounds each DNS lookup made to vet an image host.
const imageLookupTimeout = 2 * time.Second

type HTTPFetcher struct {
	client   *http.Client
	resolver Resolver
	location *time.Location
	clock    func() time.Time
}

// location is the zone assumed for feed dates that carry none, and for zone
// abbreviations that cannot be mapped to an offset. Without a client of its
// own, the fetcher builds one that refuses redirects and connections to
// internal addresses; see netguard.go.
func NewHTTPFetcher(
	client *http.Client,
	timeout time.Duration,
	location *time.Location,
) *HTTPFetcher {
	if location == nil {
		location = time.UTC
	}

	fetcher := &HTTPFetcher{
		resolver: net.DefaultResolver,
		location: location,
		clock:    time.Now,
	}

	if client == nil {
		client = newGuardedClient(
			timeout,
			func() Resolver { return fetcher.resolver },
		)
	}

	fetcher.client = client

	return fetcher
}

func (f *HTTPFetcher) Fetch(
	ctx context.Context,
	source Source,
) ([]Article, error) {
	feedURL, err := url.Parse(source.FeedURL)
	if err != nil {
		return nil, fmt.Errorf(
			"parse feed url: %w",
			err,
		)
	}

	request, err := http.NewRequestWithContext(
		withTrustedHost(ctx, feedURL),
		http.MethodGet,
		source.FeedURL,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"build request: %w",
			err,
		)
	}

	request.Header.Set("User-Agent", userAgent)
	request.Header.Set(
		"Accept",
		"application/rss+xml, application/atom+xml, application/xml;q=0.9, text/xml;q=0.8, */*;q=0.5",
	)

	response, err := f.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf(
			"request feed: %w",
			err,
		)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"feed responded with status %d",
			response.StatusCode,
		)
	}

	// Relative links resolve against where the feed was finally served from,
	// after any redirect.
	articles, err := parseFeed(
		io.LimitReader(response.Body, maxFeedSize),
		source,
		parseOptions{
			base:     response.Request.URL,
			location: f.location,
			now:      f.clock(),
		},
	)
	if err != nil {
		return nil, err
	}

	f.dropInternalImages(ctx, articles)

	return articles, nil
}

// dropInternalImages removes image URLs whose host resolves to an internal
// address, such as a public name pointing at 192.168.0.1. Parsing already
// refused internal names and literals; this catches what only DNS reveals.
//
// It is a best-effort filter, not a guarantee: the kiosk browser resolves the
// name again when it loads the image, and DNS may answer differently by then.
// Closing that gap needs the daemon to fetch images itself through the
// guarded dialer, which is the image proxy left for later.
//
// A host that cannot be resolved loses its image too; the browser could not
// load it either, and the next fetch restores it once DNS answers.
func (f *HTTPFetcher) dropInternalImages(
	ctx context.Context,
	articles []Article,
) {
	public := map[string]bool{}

	for i := range articles {
		if articles[i].ImageURL == "" {
			continue
		}

		parsed, err := url.Parse(articles[i].ImageURL)
		if err != nil {
			articles[i].ImageURL = ""

			continue
		}

		host := strings.ToLower(parsed.Hostname())

		if _, err := netip.ParseAddr(host); err == nil {
			// A literal was already checked while parsing.
			continue
		}

		allowed, seen := public[host]

		if !seen {
			allowed = f.resolvesPublicly(ctx, host)
			public[host] = allowed
		}

		if !allowed {
			articles[i].ImageURL = ""
		}
	}
}

func (f *HTTPFetcher) resolvesPublicly(
	ctx context.Context,
	host string,
) bool {
	ctx, cancel := context.WithTimeout(ctx, imageLookupTimeout)
	defer cancel()

	addrs, err := f.resolver.LookupNetIP(ctx, "ip", host)
	if err != nil || len(addrs) == 0 {
		return false
	}

	for _, addr := range addrs {
		if internalAddr(addr) {
			return false
		}
	}

	return true
}

type parseOptions struct {
	base     *url.URL
	location *time.Location
	now      time.Time
}

// The three dialects share enough element names that a single set of structs
// decodes all of them: RSS 2.0 nests items in a channel, RSS 1.0 (RDF) puts
// them beside it, and Atom calls them entries.
type xmlFeed struct {
	XMLName xml.Name
	Base    string `xml:"http://www.w3.org/XML/1998/namespace base,attr"`
	Channel struct {
		Items []xmlEntry `xml:"item"`
	} `xml:"channel"`
	Items   []xmlEntry `xml:"item"`
	Entries []xmlEntry `xml:"entry"`
}

type xmlEntry struct {
	Base         string    `xml:"http://www.w3.org/XML/1998/namespace base,attr"`
	Titles       []xmlText `xml:"title"`
	Links        []xmlLink `xml:"link"`
	GUIDs        []xmlText `xml:"guid"`
	IDs          []xmlText `xml:"id"`
	About        string    `xml:"http://www.w3.org/1999/02/22-rdf-syntax-ns# about,attr"`
	Descriptions []xmlText `xml:"description"`
	Summaries    []xmlText `xml:"summary"`
	// MediaContents must precede Contents: encoding/xml hands each element to
	// the first field that matches, and the unqualified content tag matches
	// media:content too.
	MediaContents []xmlMedia `xml:"http://search.yahoo.com/mrss/ content"`
	Contents      []xmlText  `xml:"content"`
	Encoded       string     `xml:"http://purl.org/rss/1.0/modules/content/ encoded"`
	PubDate       string     `xml:"pubDate"`
	DCDate        string     `xml:"http://purl.org/dc/elements/1.1/ date"`
	Published     string     `xml:"published"`
	Updated       string     `xml:"updated"`
	Enclosures    []xmlMedia `xml:"enclosure"`
	Thumbnails    []xmlMedia `xml:"http://search.yahoo.com/mrss/ thumbnail"`
	MediaGroups   []struct {
		Contents   []xmlMedia `xml:"http://search.yahoo.com/mrss/ content"`
		Thumbnails []xmlMedia `xml:"http://search.yahoo.com/mrss/ thumbnail"`
	} `xml:"http://search.yahoo.com/mrss/ group"`
}

// xmlText keeps the element name because an unqualified tag matches every
// namespace: media:title or media:description would otherwise overwrite the
// item's own. The raw inner XML is kept rather than the character data, since
// Atom xhtml content and descriptions with unescaped HTML arrive as child
// elements; cleanText reduces all of it to plain text.
type xmlText struct {
	XMLName xml.Name
	Inner   string `xml:",innerxml"`
}

type xmlLink struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
	Href    string `xml:"href,attr"`
	Rel     string `xml:"rel,attr"`
	Type    string `xml:"type,attr"`
}

// xmlMedia covers RSS enclosures and Media RSS elements, which name the same
// thing differently.
type xmlMedia struct {
	URL    string `xml:"url,attr"`
	Type   string `xml:"type,attr"`
	Medium string `xml:"medium,attr"`
}

const (
	namespaceRSS1 = "http://purl.org/rss/1.0/"
	namespaceAtom = "http://www.w3.org/2005/Atom"
)

func parseFeed(
	reader io.Reader,
	source Source,
	options parseOptions,
) ([]Article, error) {
	if options.location == nil {
		options.location = time.UTC
	}

	decoder := xml.NewDecoder(reader)

	// Feeds in the wild are often not well formed: HTML entities such as
	// &nbsp; in titles, unescaped ampersands, Latin-1 declared in the prolog.
	// Being lenient keeps one sloppy publisher readable. xml.HTMLAutoClose is
	// deliberately left out: it treats <link> as an empty HTML element and
	// would discard every RSS link.
	decoder.Strict = false
	decoder.Entity = xml.HTMLEntity
	decoder.CharsetReader = charsetReader

	var feed xmlFeed

	if err := decoder.Decode(&feed); err != nil {
		return nil, fmt.Errorf(
			"%w: %v",
			ErrNotAFeed,
			err,
		)
	}

	var entries []xmlEntry

	switch strings.ToLower(feed.XMLName.Local) {
	case "rss":
		entries = feed.Channel.Items

	case "rdf":
		entries = feed.Items

	case "feed":
		entries = feed.Entries

	default:
		return nil, fmt.Errorf(
			"%w: root element <%s>",
			ErrNotAFeed,
			feed.XMLName.Local,
		)
	}

	// Atom may set xml:base on the feed and again on each entry, each one
	// relative to the one above it.
	if options.base == nil {
		options.base = parseBase(httpURL(source.FeedURL))
	}

	if feed.Base != "" {
		if feedBase := parseBase(resolveURL(options.base, feed.Base)); feedBase != nil {
			options.base = feedBase
		}
	}

	articles := make([]Article, 0, len(entries))

	for _, entry := range entries {
		article, ok := normalizeEntry(entry, source, options)
		if !ok {
			continue
		}

		articles = append(articles, article)
	}

	return articles, nil
}

func normalizeEntry(
	entry xmlEntry,
	source Source,
	options parseOptions,
) (Article, bool) {
	base := options.base

	if entry.Base != "" {
		if entryBase := parseBase(resolveURL(base, entry.Base)); entryBase != nil {
			base = entryBase
		}
	}

	title := cleanText(ownText(entry.Titles))
	link := resolveURL(base, entryLink(entry))

	if title == "" || link == "" {
		return Article{}, false
	}

	guid := ownText(entry.GUIDs)

	if guid == "" {
		guid = ownText(entry.IDs)
	}

	if guid == "" {
		guid = entry.About
	}

	body := entry.Encoded

	if body == "" {
		body = ownText(entry.Contents)
	}

	description := ownText(entry.Descriptions)

	if description == "" {
		description = ownText(entry.Summaries)
	}

	if description == "" {
		description = body
	}

	// Media elements are part of the feed document and resolve against its
	// base. An image inside the article HTML was written for the article page,
	// so it resolves against that instead.
	// Image URLs are loaded by the kiosk browser on the device, so only
	// public hosts are accepted.
	imageURL := publicURL(base, entryImage(entry))

	if imageURL == "" {
		page := parseBase(link)

		imageURL = publicURL(page, firstImage(description))

		if imageURL == "" {
			imageURL = publicURL(page, firstImage(body))
		}
	}

	return Article{
		ID:           articleID(source.FeedURL, guid, link),
		Title:        title,
		Summary:      truncate(cleanText(description), maxSummaryLength),
		Source:       source.Name,
		Category:     source.Category,
		ImageURL:     imageURL,
		URL:          link,
		PublishedAt:  entryDate(entry, options.location, options.now),
		CanonicalURL: normalizeURL(link),
	}, true
}

// ownText returns the first value from the feed's own vocabulary, skipping
// same-named elements from extension namespaces.
func ownText(values []xmlText) string {
	for _, value := range values {
		switch value.XMLName.Space {
		case "", namespaceRSS1, namespaceAtom:
		default:
			continue
		}

		if text := strings.TrimSpace(value.Inner); text != "" {
			return text
		}
	}

	return ""
}

func entryLink(entry xmlEntry) string {
	for _, link := range entry.Links {
		switch link.XMLName.Space {
		case "", namespaceRSS1, namespaceAtom:
		default:
			continue
		}

		// RSS carries the link as text. Atom carries it as an attribute, and
		// lists other relations — the feed itself, comments, an enclosure —
		// beside the article.
		if value := strings.TrimSpace(link.Value); value != "" {
			return value
		}

		if link.Href != "" && (link.Rel == "" || link.Rel == "alternate") {
			return link.Href
		}
	}

	return ""
}

func entryImage(entry xmlEntry) string {
	for _, group := range entry.MediaGroups {
		if url := mediaImage(group.Contents); url != "" {
			return url
		}

		if url := mediaImage(group.Thumbnails); url != "" {
			return url
		}
	}

	if url := mediaImage(entry.Enclosures); url != "" {
		return url
	}

	if url := mediaImage(entry.MediaContents); url != "" {
		return url
	}

	if url := mediaImage(entry.Thumbnails); url != "" {
		return url
	}

	for _, link := range entry.Links {
		if link.Rel == "enclosure" && strings.HasPrefix(link.Type, "image/") {
			return link.Href
		}
	}

	return ""
}

// mediaImage picks the first entry that is an image, or that does not say what
// it is: thumbnails rarely carry a type.
func mediaImage(media []xmlMedia) string {
	for _, item := range media {
		if item.URL == "" {
			continue
		}

		if item.Type != "" && !strings.HasPrefix(item.Type, "image/") {
			continue
		}

		if item.Medium != "" && item.Medium != "image" {
			continue
		}

		return item.URL
	}

	return ""
}

var imagePattern = regexp.MustCompile(
	`(?i)<img[^>]+src\s*=\s*["']([^"']+)["']`,
)

// firstImage finds an inline image in HTML content, which is where many
// Brazilian portals put the article photo instead of an enclosure.
func firstImage(content string) string {
	match := imagePattern.FindStringSubmatch(html.UnescapeString(content))
	if match == nil {
		return ""
	}

	return html.UnescapeString(match[1])
}

// zonedLayouts carry their own offset. Zone abbreviations never reach them:
// parseDate rewrites a known one as an offset first, because time.Parse would
// otherwise read an abbreviation it does not know, such as BRT, as UTC.
var zonedLayouts = []string{
	time.RFC3339Nano,
	"Mon, 2 Jan 2006 15:04:05 -0700",
	"Mon, 2 Jan 2006 15:04:05 -07:00",
	"Mon, 2 Jan 2006 15:04 -0700",
	"2 Jan 2006 15:04:05 -0700",
	"2 Jan 2006 15:04 -0700",
	"2 Jan 06 15:04 -0700",
	"2006-01-02T15:04:05-0700",
	"2006-01-02 15:04:05 -0700",
}

// localLayouts carry no zone and are read in the configured location.
var localLayouts = []string{
	"Mon, 2 Jan 2006 15:04:05",
	"Mon, 2 Jan 2006 15:04",
	"2 Jan 2006 15:04:05",
	"2006-01-02T15:04:05",
	"2006-01-02T15:04",
	"2006-01-02 15:04:05",
	"2006-01-02",
}

// zoneOffsets maps the abbreviations found in feed dates to offsets. RFC 822
// defines the North American and military ones; the Brazilian ones are what
// local portals emit. Ambiguous abbreviations such as IST or CST outside the
// RFC list are left out on purpose and fall back to the configured location.
var zoneOffsets = map[string]string{
	"UT":   "+0000",
	"UTC":  "+0000",
	"GMT":  "+0000",
	"Z":    "+0000",
	"EST":  "-0500",
	"EDT":  "-0400",
	"CST":  "-0600",
	"CDT":  "-0500",
	"MST":  "-0700",
	"MDT":  "-0600",
	"PST":  "-0800",
	"PDT":  "-0700",
	"BRT":  "-0300",
	"BRST": "-0200",
	"AMT":  "-0400",
	"FNT":  "-0200",
	"WET":  "+0000",
	"WEST": "+0100",
	"CET":  "+0100",
	"CEST": "+0200",
}

// parseDate reads the date formats feeds use in practice. A trailing zone
// abbreviation is replaced by its offset when it is known and dropped when it
// is not, in which case the date is read in location like a date with no zone.
func parseDate(value string, location *time.Location) (time.Time, bool) {
	value = strings.Join(strings.Fields(value), " ")

	if value == "" {
		return time.Time{}, false
	}

	if index := strings.LastIndex(value, " "); index > 0 {
		zone := value[index+1:]

		if isZoneAbbreviation(zone) {
			if offset, ok := zoneOffsets[strings.ToUpper(zone)]; ok {
				value = value[:index+1] + offset
			} else {
				value = value[:index]
			}
		}
	}

	for _, layout := range zonedLayouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, true
		}
	}

	for _, layout := range localLayouts {
		if parsed, err := time.ParseInLocation(layout, value, location); err == nil {
			return parsed, true
		}
	}

	return time.Time{}, false
}

func isZoneAbbreviation(value string) bool {
	if value == "" || len(value) > 5 {
		return false
	}

	for _, r := range value {
		if (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') {
			return false
		}
	}

	return true
}

// entryDate returns the publication time, or the zero time when the feed gives
// none, so the repository can fall back to when the article was first seen.
// A date in the future is clamped: feeds with a wrong timezone would otherwise
// pin their stories to the top of the slideshow for hours.
func entryDate(
	entry xmlEntry,
	location *time.Location,
	now time.Time,
) time.Time {
	for _, value := range []string{
		entry.PubDate,
		entry.Published,
		entry.DCDate,
		entry.Updated,
	} {
		parsed, ok := parseDate(value, location)
		if !ok {
			continue
		}

		if parsed.After(now) {
			return now
		}

		return parsed
	}

	return time.Time{}
}

var (
	tagPattern             = regexp.MustCompile(`<[^>]*>`)
	spaceBeforePunctuation = regexp.MustCompile(` ([.,;:!?)])`)
	cdataMarkers           = strings.NewReplacer("<![CDATA[", "", "]]>", "")
)

// cleanText turns feed HTML into plain text. Unescaping runs before and after
// stripping tags because feeds escape their HTML a varying number of times.
// Tags become spaces so block elements do not glue words together, and the
// space that leaves before punctuation after an inline tag is removed again.
func cleanText(value string) string {
	value = cdataMarkers.Replace(value)
	value = html.UnescapeString(value)
	value = tagPattern.ReplaceAllString(value, " ")
	value = html.UnescapeString(value)

	// strings.Fields also splits on the non-breaking spaces &nbsp; becomes.
	value = strings.Join(strings.Fields(value), " ")

	return spaceBeforePunctuation.ReplaceAllString(value, "$1")
}

func truncate(value string, limit int) string {
	if utf8.RuneCountInString(value) <= limit {
		return value
	}

	runes := []rune(value)[:limit]
	cut := string(runes)

	if index := strings.LastIndex(cut, " "); index > limit/2 {
		cut = cut[:index]
	}

	return strings.TrimRight(cut, " ,.;:-") + "…"
}

// charsetReader decodes the single-byte encodings still common in Brazilian
// feeds. encoding/xml only understands UTF-8 on its own.
func charsetReader(
	charset string,
	input io.Reader,
) (io.Reader, error) {
	switch strings.ToLower(strings.TrimSpace(charset)) {
	case "utf-8", "utf8", "us-ascii", "ascii":
		return input, nil

	case "iso-8859-1", "iso8859-1", "latin1", "latin-1",
		"windows-1252", "cp1252", "iso-8859-15":
		data, err := io.ReadAll(input)
		if err != nil {
			return nil, err
		}

		var builder strings.Builder

		builder.Grow(len(data))

		for _, b := range data {
			if b >= 0x80 && b <= 0x9f {
				builder.WriteRune(windows1252[b-0x80])

				continue
			}

			builder.WriteRune(rune(b))
		}

		return strings.NewReader(builder.String()), nil
	}

	return nil, fmt.Errorf(
		"unsupported charset %q",
		charset,
	)
}

// windows1252 maps the bytes 0x80 to 0x9f, where Windows-1252 puts curly
// quotes and dashes and Latin-1 has control characters. Feeds declaring
// ISO-8859-1 are usually Windows-1252 in practice.
var windows1252 = [32]rune{
	'€', '\u0081', '‚', 'ƒ', '„', '…', '†', '‡',
	'ˆ', '‰', 'Š', '‹', 'Œ', '\u008d', 'Ž', '\u008f',
	'\u0090', '‘', '’', '“', '”', '•', '–', '—',
	'˜', '™', 'š', '›', 'œ', '\u009d', 'ž', 'Ÿ',
}
