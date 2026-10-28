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

var saoPaulo = mustLoadLocation("America/Sao_Paulo")

func mustLoadLocation(name string) *time.Location {
	location, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}

	return location
}

func testOptions() parseOptions {
	return parseOptions{
		base:     parseBase(testSource.FeedURL),
		location: saoPaulo,
		now:      testNow,
	}
}

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

	articles, err := parseFeed(strings.NewReader(feed), testSource, testOptions())
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

	articles, err := parseFeed(strings.NewReader(feed), testSource, testOptions())
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

	articles, err := parseFeed(strings.NewReader(feed), testSource, testOptions())
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
		testOptions(),
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

	articles, err := parseFeed(strings.NewReader(feed), testSource, testOptions())
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

func TestArticleIDIsScopedToTheFeed(t *testing.T) {
	link := "https://Example.com/x/?utm_source=a#top"

	sameFeed := articleID("https://a/feed", "", "https://example.com/x")
	first := articleID("https://a/feed", "", link)
	second := articleID("https://b/feed", "", link)

	if first == "" || first != sameFeed {
		t.Fatalf("expected the normalised link to give the same id, got %q and %q", first, sameFeed)
	}

	if first == second {
		t.Fatalf("expected different ids for the same link in two feeds")
	}

	if articleID("https://a/feed", "1", "") == articleID("https://b/feed", "1", "") {
		t.Fatalf("expected guids to be scoped to the feed")
	}
}

func TestHTTPFetcherReportsStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "gone", http.StatusNotFound)
		},
	))
	defer server.Close()

	fetcher := NewHTTPFetcher(nil, time.Second, time.UTC)

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

func TestParseResolvesRelativeURLs(t *testing.T) {
	const feed = `<rss xmlns:media="http://search.yahoo.com/mrss/"><channel>
		<item>
			<title>Relativo</title>
			<link>/noticias/relativo</link>
			<description><![CDATA[<img src="../img/foto.jpg"> Texto]]></description>
		</item>
		<item>
			<title>Protocolo</title>
			<link>https://example.com/noticias/protocolo</link>
			<media:thumbnail url="//cdn.example.com/thumb.jpg"/>
		</item>
		<item>
			<title>Inseguro</title>
			<link>https://example.com/noticias/inseguro</link>
			<description><![CDATA[<img src="javascript:alert(1)"> <img src="data:image/png;base64,AAAA">]]></description>
		</item>
	</channel></rss>`

	articles, err := parseFeed(strings.NewReader(feed), testSource, testOptions())
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if len(articles) != 3 {
		t.Fatalf("expected 3 articles, got %d", len(articles))
	}

	if articles[0].URL != "https://example.com/noticias/relativo" {
		t.Errorf("link = %q", articles[0].URL)
	}

	// Inline HTML belongs to the article page, so it resolves against that.
	if articles[0].ImageURL != "https://example.com/img/foto.jpg" {
		t.Errorf("inline image = %q", articles[0].ImageURL)
	}

	if articles[1].ImageURL != "https://cdn.example.com/thumb.jpg" {
		t.Errorf("protocol-relative image = %q", articles[1].ImageURL)
	}

	if articles[2].ImageURL != "" {
		t.Errorf("unsafe image kept: %q", articles[2].ImageURL)
	}
}

func TestParseAtomXMLBase(t *testing.T) {
	const feed = `<feed xmlns="http://www.w3.org/2005/Atom" xml:base="https://blog.example.org/">
		<entry xml:base="posts/">
			<id>1</id>
			<title>Base</title>
			<link href="primeiro"/>
		</entry>
	</feed>`

	articles, err := parseFeed(strings.NewReader(feed), testSource, testOptions())
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if len(articles) != 1 || articles[0].URL != "https://blog.example.org/posts/primeiro" {
		t.Fatalf("articles = %+v", articles)
	}
}

func TestHTTPFetcherResolvesAgainstTheFinalURL(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/old", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/new/feed.xml", http.StatusMovedPermanently)
	})

	mux.HandleFunc("/new/feed.xml", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<rss><channel><item><title>T</title><link>artigo</link></item></channel></rss>`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	articles, err := NewHTTPFetcher(nil, time.Second, time.UTC).Fetch(
		context.Background(),
		Source{FeedURL: server.URL + "/old"},
	)
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}

	if len(articles) != 1 || articles[0].URL != server.URL+"/new/artigo" {
		t.Fatalf("articles = %+v", articles)
	}
}

func TestParseDate(t *testing.T) {
	cases := []struct {
		value string
		want  time.Time
	}{
		{"Tue, 27 Oct 2026 18:30:00 +0000", time.Date(2026, 10, 27, 18, 30, 0, 0, time.UTC)},
		{"Tue, 27 Oct 2026 18:30:00 GMT", time.Date(2026, 10, 27, 18, 30, 0, 0, time.UTC)},
		{"Tue, 27 Oct 2026 15:30:00 BRT", time.Date(2026, 10, 27, 18, 30, 0, 0, time.UTC)},
		{"Tue, 27 Oct 2026 14:30:00 EDT", time.Date(2026, 10, 27, 18, 30, 0, 0, time.UTC)},
		{"Tue, 27 Oct 2026 15:30:00 -03:00", time.Date(2026, 10, 27, 18, 30, 0, 0, time.UTC)},
		{"27 Oct 26 18:30 Z", time.Date(2026, 10, 27, 18, 30, 0, 0, time.UTC)},
		{"2026-10-27T18:30:00Z", time.Date(2026, 10, 27, 18, 30, 0, 0, time.UTC)},
		{"2026-10-27T15:30:00.123-03:00", time.Date(2026, 10, 27, 18, 30, 0, 123000000, time.UTC)},

		// No zone, and a zone nobody can map: read in the configured location.
		{"2026-10-27 15:30:00", time.Date(2026, 10, 27, 18, 30, 0, 0, time.UTC)},
		{"Tue, 27 Oct 2026 15:30:00", time.Date(2026, 10, 27, 18, 30, 0, 0, time.UTC)},
		{"Tue, 27 Oct 2026 15:30:00 XYZ", time.Date(2026, 10, 27, 18, 30, 0, 0, time.UTC)},
		{"2026-10-27", time.Date(2026, 10, 27, 3, 0, 0, 0, time.UTC)},
	}

	for _, c := range cases {
		got, ok := parseDate(c.value, saoPaulo)

		if !ok {
			t.Errorf("%q not parsed", c.value)

			continue
		}

		if !got.Equal(c.want) {
			t.Errorf("%q = %v, want %v", c.value, got.UTC(), c.want)
		}
	}

	if _, ok := parseDate("ontem", saoPaulo); ok {
		t.Errorf("expected garbage to be rejected")
	}
}
