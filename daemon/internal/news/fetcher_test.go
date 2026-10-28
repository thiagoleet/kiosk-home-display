package news

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var testSource = Source{
	ID:       1,
	Name:     "Fonte",
	FeedURL:  "https://example.com/feed",
	Category: "science",
}

var testNow = time.Date(2026, 10, 27, 19, 0, 0, 0, time.UTC)

func TestParseRSS2(t *testing.T) {
	const feed = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:media="http://search.yahoo.com/mrss/" xmlns:atom="http://www.w3.org/2005/Atom">
<channel>
	<title>Fonte</title>
	<atom:link href="https://example.com/feed" rel="self"/>
	<item>
		<title>Nova descoberta &amp; pesquisa</title>
		<link>https://example.com/article?utm_source=rss</link>
		<guid>article-123</guid>
		<description><![CDATA[<p>Pesquisadores&nbsp;anunciaram <b>algo</b>.</p>]]></description>
		<pubDate>Tue, 27 Oct 2026 18:30:00 +0000</pubDate>
		<media:title>Ignored media title</media:title>
		<media:content url="https://example.com/image.jpg" medium="image"/>
	</item>
	<item>
		<title>Sem link</title>
	</item>
</channel>
</rss>`

	articles, err := parseFeed(strings.NewReader(feed), testSource, testNow)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if len(articles) != 1 {
		t.Fatalf("expected 1 article, got %d", len(articles))
	}

	article := articles[0]

	if article.Title != "Nova descoberta & pesquisa" {
		t.Errorf("title = %q", article.Title)
	}

	if article.Summary != "Pesquisadores anunciaram algo." {
		t.Errorf("summary = %q", article.Summary)
	}

	if article.ImageURL != "https://example.com/image.jpg" {
		t.Errorf("image = %q", article.ImageURL)
	}

	if article.Source != "Fonte" || article.Category != "science" {
		t.Errorf("source/category = %q/%q", article.Source, article.Category)
	}

	want := time.Date(2026, 10, 27, 18, 30, 0, 0, time.UTC)

	if !article.PublishedAt.Equal(want) {
		t.Errorf("published = %v", article.PublishedAt)
	}

	if article.ID != articleID(testSource.FeedURL, "article-123", "") {
		t.Errorf("id not derived from guid")
	}
}

func TestParseAtom(t *testing.T) {
	const feed = `<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
	<title>Fonte</title>
	<entry>
		<id>tag:example.com,2026:1</id>
		<title type="html">T&#237;tulo</title>
		<link rel="replies" href="https://example.com/comments"/>
		<link rel="alternate" href="https://example.com/atom-article"/>
		<link rel="enclosure" type="image/png" href="https://example.com/a.png"/>
		<summary>Resumo</summary>
		<updated>2026-10-27T10:00:00Z</updated>
		<published>2026-10-27T09:00:00-03:00</published>
	</entry>
</feed>`

	articles, err := parseFeed(strings.NewReader(feed), testSource, testNow)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if len(articles) != 1 {
		t.Fatalf("expected 1 article, got %d", len(articles))
	}

	article := articles[0]

	if article.Title != "Título" {
		t.Errorf("title = %q", article.Title)
	}

	if article.URL != "https://example.com/atom-article" {
		t.Errorf("url = %q", article.URL)
	}

	if article.ImageURL != "https://example.com/a.png" {
		t.Errorf("image = %q", article.ImageURL)
	}

	if !article.PublishedAt.Equal(time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)) {
		t.Errorf("published = %v", article.PublishedAt)
	}
}

func TestParseRDFInLatin1(t *testing.T) {
	// "Notícia" encoded as ISO-8859-1, with a Windows-1252 right quote.
	feed := "<?xml version=\"1.0\" encoding=\"ISO-8859-1\"?>\n" +
		`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#" xmlns="http://purl.org/rss/1.0/" xmlns:dc="http://purl.org/dc/elements/1.1/">` +
		`<channel rdf:about="https://example.com/"><title>Fonte</title></channel>` +
		`<item rdf:about="https://example.com/rdf-article">` +
		"<title>Not\xedcia \x93boa\x94</title>" +
		`<link>https://example.com/rdf-article</link>` +
		`<description>Texto com <img src="https://example.com/inline.jpg"> imagem</description>` +
		`<dc:date>2026-10-26T08:00:00Z</dc:date>` +
		`</item></rdf:RDF>`

	articles, err := parseFeed(strings.NewReader(feed), testSource, testNow)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if len(articles) != 1 {
		t.Fatalf("expected 1 article, got %d", len(articles))
	}

	if articles[0].Title != "Notícia “boa”" {
		t.Errorf("title = %q", articles[0].Title)
	}

	if articles[0].ImageURL != "https://example.com/inline.jpg" {
		t.Errorf("image = %q", articles[0].ImageURL)
	}

	if articles[0].PublishedAt.IsZero() {
		t.Errorf("dc:date not parsed")
	}
}

func TestParseRejectsHTML(t *testing.T) {
	_, err := parseFeed(
		strings.NewReader(`<html><body>Not found</body></html>`),
		testSource,
		testNow,
	)

	if !errors.Is(err, ErrNotAFeed) {
		t.Fatalf("expected ErrNotAFeed, got %v", err)
	}
}

func TestParseClampsFutureDatesAndDropsUnsafeLinks(t *testing.T) {
	const feed = `<rss><channel>
		<item><title>Futuro</title><link>https://example.com/f</link><pubDate>Fri, 30 Oct 2026 10:00:00 +0000</pubDate></item>
		<item><title>Script</title><link>javascript:alert(1)</link></item>
	</channel></rss>`

	articles, err := parseFeed(strings.NewReader(feed), testSource, testNow)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if len(articles) != 1 {
		t.Fatalf("expected 1 article, got %d", len(articles))
	}

	if !articles[0].PublishedAt.Equal(testNow) {
		t.Errorf("future date not clamped: %v", articles[0].PublishedAt)
	}
}

func TestArticleIDFallsBackToNormalizedURL(t *testing.T) {
	first := articleID("https://a/feed", "", "https://Example.com/x/?utm_source=a#top")
	second := articleID("https://b/feed", "", "https://example.com/x")

	if first == "" || first != second {
		t.Fatalf("expected equal ids, got %q and %q", first, second)
	}
}

func TestHTTPFetcherReportsStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "gone", http.StatusNotFound)
		},
	))
	defer server.Close()

	fetcher := NewHTTPFetcher(nil, time.Second)

	_, err := fetcher.Fetch(
		context.Background(),
		Source{FeedURL: server.URL},
	)

	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("expected a 404 error, got %v", err)
	}
}

func TestParseSources(t *testing.T) {
	sources, err := ParseSources(
		"G1|https://g1.globo.com/rss/g1/; NASA | https://www.nasa.gov/feed/ | science ;",
	)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if len(sources) != 2 {
		t.Fatalf("expected 2 sources, got %d", len(sources))
	}

	if sources[0].Category != DefaultCategory || sources[1].Category != "science" {
		t.Errorf("categories = %q, %q", sources[0].Category, sources[1].Category)
	}

	for _, invalid := range []string{
		"no-url",
		"Name|ftp://example.com/feed",
		"|https://example.com/feed",
		"A|https://example.com/feed;B|https://example.com/feed",
	} {
		if _, err := ParseSources(invalid); err == nil {
			t.Errorf("expected %q to be rejected", invalid)
		}
	}
}
