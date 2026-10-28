package news

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPublicHost(t *testing.T) {
	internal := []string{
		// IPv4 and the spellings browsers also read as IPv4.
		"127.0.0.1",
		"127.1",
		"2130706433",
		"0x7f000001",
		"0x7f.0.0.1",
		"0177.0.0.1",
		"10.1.2.3",
		"172.16.0.1",
		"172.31.255.255",
		"192.168.1.1",
		"169.254.169.254",
		"100.64.0.1",
		"0.0.0.0",
		"0",
		"255.255.255.255",
		"224.0.0.1",
		"192.0.2.10",
		"1.2.3.256",
		"1.2.3.4.5",

		// IPv6.
		"::1",
		"::",
		"fe80::1",
		"fe80::1%eth0",
		"fc00::1",
		"fd12:3456::1",
		"::ffff:127.0.0.1",
		"::ffff:10.0.0.1",
		"::ffff:7f00:1",
		"::127.0.0.1",
		"64:ff9b::a00:1",
		"2002:c0a8:101::1",
		"ff02::1",
		"fec0::1",
		"2001:db8::1",

		// Names that only resolve locally.
		"localhost",
		"LOCALHOST.",
		"app.localhost",
		"printer.local",
		"router",
		"nas.lan",
		"metadata.internal",
		"box.home.arpa",
		"",
	}

	public := []string{
		"8.8.8.8",
		"1.1.1.1",
		"172.32.0.1",
		"93.184.216.34",
		"2606:4700:4700::1111",
		"2002:808:808::1",
		"example.com",
		"cdn.example.com.",
		"s2-g1.glbimg.com",
	}

	for _, host := range internal {
		if publicHost(host) {
			t.Errorf("%q accepted as public", host)
		}
	}

	for _, host := range public {
		if !publicHost(host) {
			t.Errorf("%q rejected", host)
		}
	}
}

func TestPublicURL(t *testing.T) {
	page := parseBase("https://news.example.com/2026/article")

	rejected := []string{
		"http://127.0.0.1/a.png",
		"http://[::1]/a.png",
		"http://[fe80::1%25eth0]/a.png",
		"http://[::ffff:192.168.0.1]/a.png",
		"http://2130706433/a.png",
		"http://0x7f.1/a.png",
		"http://localhost:8080/api/display/sleep",
		"//localhost/a.png",
		"http://router/admin.png",
		"http://user@cdn.example.com/a.png",
		"http://cdn.example.com\\@127.0.0.1/a.png",
		"javascript:alert(1)",
		"data:image/png;base64,AAAA",
		"file:///etc/passwd",
		"ftp://cdn.example.com/a.png",
	}

	for _, raw := range rejected {
		if got := publicURL(page, raw); got != "" {
			t.Errorf("%q accepted as %q", raw, got)
		}
	}

	accepted := map[string]string{
		"https://cdn.example.com/a.png": "https://cdn.example.com/a.png",
		"//cdn.example.com/a.png":       "https://cdn.example.com/a.png",
		"/img/a.png":                    "https://news.example.com/img/a.png",
		"http://[2606:4700::1]/a.png":   "http://[2606:4700::1]/a.png",
	}

	for raw, want := range accepted {
		if got := publicURL(page, raw); got != want {
			t.Errorf("%q = %q, want %q", raw, got, want)
		}
	}
}

func TestParseDropsInternalImages(t *testing.T) {
	const feed = `<rss xmlns:media="http://search.yahoo.com/mrss/"><channel>
		<item>
			<title>Media</title>
			<link>https://example.com/a</link>
			<media:content url="http://192.168.0.1/cam.jpg" medium="image"/>
		</item>
		<item>
			<title>Inline</title>
			<link>https://example.com/b</link>
			<description><![CDATA[<img src="http://localhost:8080/x.png">]]></description>
		</item>
		<item>
			<title>Fallback</title>
			<link>https://example.com/c</link>
			<enclosure url="http://169.254.169.254/latest" type="image/png"/>
			<description><![CDATA[<img src="https://cdn.example.com/ok.png">]]></description>
		</item>
	</channel></rss>`

	articles, err := parseFeed(strings.NewReader(feed), testSource, testOptions())
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if articles[0].ImageURL != "" || articles[1].ImageURL != "" {
		t.Errorf("internal images kept: %q, %q", articles[0].ImageURL, articles[1].ImageURL)
	}

	// An internal media element does not hide a public inline image.
	if articles[2].ImageURL != "https://cdn.example.com/ok.png" {
		t.Errorf("fallback image = %q", articles[2].ImageURL)
	}
}

// stubResolver answers from a table, optionally changing its answer after the
// first lookup of a host to model DNS rebinding.
type stubResolver struct {
	mu      sync.Mutex
	answers map[string][]string
	later   map[string][]string
	calls   map[string]int
}

func newStubResolver() *stubResolver {
	return &stubResolver{
		answers: map[string][]string{},
		later:   map[string][]string{},
		calls:   map[string]int{},
	}
}

func (r *stubResolver) LookupNetIP(
	ctx context.Context,
	network string,
	host string,
) ([]netip.Addr, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.calls[host]++

	answer, ok := r.answers[host]

	if later, changes := r.later[host]; changes && r.calls[host] > 1 {
		answer = later
	}

	if !ok {
		return nil, errors.New("no such host")
	}

	addrs := make([]netip.Addr, 0, len(answer))

	for _, value := range answer {
		addrs = append(addrs, netip.MustParseAddr(value))
	}

	return addrs, nil
}

func (r *stubResolver) count(host string) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.calls[host]
}

func newTestFetcher(resolver Resolver) *HTTPFetcher {
	fetcher := NewHTTPFetcher(nil, 2*time.Second, time.UTC)
	fetcher.resolver = resolver

	return fetcher
}

// redirectingFeed serves a redirect to target from /feed.
func redirectingFeed(t *testing.T, target string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, target, http.StatusFound)
		},
	))

	t.Cleanup(server.Close)

	return server
}

func TestRedirectToAnotherLocalServerIsRefused(t *testing.T) {
	var reached atomic.Int32

	internal := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			reached.Add(1)
		},
	))
	defer internal.Close()

	// The configured feed is local, which is allowed; where it redirects to
	// is another port on loopback, which is not.
	feed := redirectingFeed(t, internal.URL+"/api/display/sleep")

	_, err := newTestFetcher(newStubResolver()).Fetch(
		context.Background(),
		Source{FeedURL: feed.URL + "/feed"},
	)

	if !errors.Is(err, ErrInternalAddress) {
		t.Fatalf("expected ErrInternalAddress, got %v", err)
	}

	if reached.Load() != 0 {
		t.Fatalf("the internal server was reached")
	}
}

func TestRedirectsToInternalOrUnsafeTargetsAreRefused(t *testing.T) {
	resolver := newStubResolver()
	resolver.answers["private.example.com"] = []string{"10.0.0.5"}
	resolver.answers["mixed.example.com"] = []string{"93.184.216.34", "127.0.0.1"}
	resolver.answers["v6.example.com"] = []string{"::ffff:192.168.1.1"}

	targets := map[string]error{
		"http://localhost:9/feed":           ErrInternalAddress,
		"http://169.254.169.254/latest":     ErrInternalAddress,
		"http://[::1]:9/feed":               ErrInternalAddress,
		"http://2130706433/feed":            ErrInternalAddress,
		"http://router/feed":                ErrInternalAddress,
		"http://private.example.com/feed":   ErrInternalAddress,
		"http://mixed.example.com/feed":     ErrInternalAddress,
		"http://v6.example.com/feed":        ErrInternalAddress,
		"file:///etc/passwd":                nil,
		"ftp://example.com/feed":            nil,
		"http://user:pass@example.com/feed": nil,
		"https://unresolvable.example/feed": nil,
	}

	for target, want := range targets {
		t.Run(target, func(t *testing.T) {
			feed := redirectingFeed(t, target)

			_, err := newTestFetcher(resolver).Fetch(
				context.Background(),
				Source{FeedURL: feed.URL},
			)

			if err == nil {
				t.Fatalf("redirect to %s followed", target)
			}

			if want != nil && !errors.Is(err, want) {
				t.Fatalf("got %v, want %v", err, want)
			}
		})
	}
}

func TestDialGuardChecksTheAddressItConnectsTo(t *testing.T) {
	// The redirect check sees a public name; DNS then answers with loopback
	// when the connection is made. The guard resolves at dial time and
	// refuses, instead of trusting the first answer.
	resolver := newStubResolver()
	resolver.answers["rebind.example.com"] = []string{"93.184.216.34"}
	resolver.later["rebind.example.com"] = []string{"127.0.0.1"}

	// Prime the first answer as an earlier lookup would.
	if _, err := resolver.LookupNetIP(context.Background(), "ip", "rebind.example.com"); err != nil {
		t.Fatalf("prime: %v", err)
	}

	feed := redirectingFeed(t, "http://rebind.example.com/feed")

	_, err := newTestFetcher(resolver).Fetch(
		context.Background(),
		Source{FeedURL: feed.URL},
	)

	if !errors.Is(err, ErrInternalAddress) {
		t.Fatalf("expected ErrInternalAddress, got %v", err)
	}
}

func TestTooManyRedirects(t *testing.T) {
	var server *httptest.Server

	server = httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, server.URL+r.URL.Path+"x", http.StatusFound)
		},
	))
	defer server.Close()

	_, err := newTestFetcher(newStubResolver()).Fetch(
		context.Background(),
		Source{FeedURL: server.URL + "/r"},
	)

	if err == nil || !strings.Contains(err.Error(), "redirects") {
		t.Fatalf("expected a redirect limit error, got %v", err)
	}
}

func TestFetchDropsImagesResolvingInternally(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`<rss xmlns:media="http://search.yahoo.com/mrss/"><channel>
				<item><title>1</title><link>https://example.com/1</link>
					<media:thumbnail url="https://cam.example.net/1.jpg"/></item>
				<item><title>2</title><link>https://example.com/2</link>
					<media:thumbnail url="https://cam.example.net/2.jpg"/></item>
				<item><title>3</title><link>https://example.com/3</link>
					<media:thumbnail url="https://cdn.example.org/3.jpg"/></item>
				<item><title>4</title><link>https://example.com/4</link>
					<media:thumbnail url="https://gone.example.org/4.jpg"/></item>
			</channel></rss>`))
		},
	))
	defer server.Close()

	resolver := newStubResolver()
	resolver.answers["cam.example.net"] = []string{"192.168.0.20"}
	resolver.answers["cdn.example.org"] = []string{"151.101.1.1", "2a04:4e42::1"}

	articles, err := newTestFetcher(resolver).Fetch(
		context.Background(),
		Source{FeedURL: server.URL},
	)
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}

	got := []string{}

	for _, article := range articles {
		got = append(got, article.ImageURL)
	}

	want := []string{"", "", "https://cdn.example.org/3.jpg", ""}

	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("images = %q, want %q", got, want)
	}

	// One lookup per host per fetch.
	if resolver.count("cam.example.net") != 1 {
		t.Errorf("cam.example.net looked up %d times", resolver.count("cam.example.net"))
	}
}
