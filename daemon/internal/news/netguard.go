package news

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ErrInternalAddress is returned when a redirect, or a connection made on
// behalf of one, would reach the device itself or the local network.
var ErrInternalAddress = errors.New("refusing to reach an internal address")

// maxRedirects bounds how many hops a feed may take before it is served.
const maxRedirects = 5

// Resolver is the part of net.Resolver the guard uses. Tests replace it.
type Resolver interface {
	LookupNetIP(ctx context.Context, network, host string) ([]netip.Addr, error)
}

// blockedPrefixes are the ranges netip's own predicates leave out: shared
// address space, benchmarking, documentation and reserved blocks, and the
// IPv6 transition ranges that can tunnel to an arbitrary IPv4 address.
var blockedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("::/96"),
	netip.MustParsePrefix("64:ff9b::/96"),
	netip.MustParsePrefix("64:ff9b:1::/48"),
	netip.MustParsePrefix("100::/64"),
	netip.MustParsePrefix("2001::/32"),
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("fec0::/10"),
}

var sixToFour = netip.MustParsePrefix("2002::/16")

// internalAddr reports whether an address belongs to the device or a private
// network rather than to the public internet.
func internalAddr(addr netip.Addr) bool {
	addr = addr.Unmap().WithZone("")

	if !addr.IsValid() ||
		addr.IsLoopback() ||
		addr.IsPrivate() ||
		addr.IsLinkLocalUnicast() ||
		addr.IsLinkLocalMulticast() ||
		addr.IsInterfaceLocalMulticast() ||
		addr.IsMulticast() ||
		addr.IsUnspecified() {
		return true
	}

	for _, prefix := range blockedPrefixes {
		if prefix.Contains(addr) {
			return true
		}
	}

	// 6to4 embeds an IPv4 address in bytes 2 to 5; it is only as public as
	// the address it carries.
	if sixToFour.Contains(addr) {
		bytes := addr.As16()

		return internalAddr(netip.AddrFrom4([4]byte{
			bytes[2], bytes[3], bytes[4], bytes[5],
		}))
	}

	return false
}

// localSuffixes are names that only resolve inside a home or office network.
var localSuffixes = []string{
	".localhost",
	".local",
	".localdomain",
	".internal",
	".intranet",
	".lan",
	".home",
	".home.arpa",
	".corp",
	".private",
}

// publicHost decides from the name alone whether a host can be on the public
// internet. It reads IP literals the way a browser does, including the
// decimal, octal and hexadecimal IPv4 spellings, because the kiosk browser,
// not Go, is what eventually loads an image URL.
//
// A name that passes may still resolve to a private address; that needs a
// DNS lookup, which the dial guard and the image check make separately.
func publicHost(host string) bool {
	host = strings.ToLower(strings.TrimSuffix(host, "."))

	if host == "" {
		return false
	}

	if addr, err := netip.ParseAddr(host); err == nil {
		return !internalAddr(addr)
	}

	if addr, numeric, ok := browserIPv4(host); numeric {
		return ok && !internalAddr(addr)
	}

	// A single label is completed by the local search domain: router, nas,
	// localhost.
	if !strings.Contains(host, ".") {
		return false
	}

	for _, suffix := range localSuffixes {
		if strings.HasSuffix(host, suffix) {
			return false
		}
	}

	return true
}

// browserIPv4 follows the WHATWG URL host parser. A host whose last label is
// a number is an IPv4 address to a browser, written with one to four parts in
// decimal, octal (leading 0) or hexadecimal (0x): 2130706433, 0x7f.1 and
// 0177.0.0.1 all mean 127.0.0.1. numeric reports that the host is meant as an
// address; ok that it is a valid one. A numeric host that is not valid is
// rejected by browsers, and here.
func browserIPv4(host string) (addr netip.Addr, numeric bool, ok bool) {
	parts := strings.Split(host, ".")

	if !numericPart(parts[len(parts)-1]) {
		return netip.Addr{}, false, false
	}

	if len(parts) > 4 {
		return netip.Addr{}, true, false
	}

	values := make([]uint64, len(parts))

	for i, part := range parts {
		value, valid := parseIPv4Part(part)
		if !valid {
			return netip.Addr{}, true, false
		}

		values[i] = value
	}

	var result uint64

	for i, value := range values[:len(values)-1] {
		if value > 255 {
			return netip.Addr{}, true, false
		}

		result |= value << (8 * (3 - i))
	}

	last := values[len(values)-1]

	if last >= 1<<(8*(5-len(values))) {
		return netip.Addr{}, true, false
	}

	result |= last

	return netip.AddrFrom4([4]byte{
		byte(result >> 24),
		byte(result >> 16),
		byte(result >> 8),
		byte(result),
	}), true, true
}

func numericPart(part string) bool {
	if part == "" {
		return false
	}

	if strings.HasPrefix(part, "0x") || strings.HasPrefix(part, "0X") {
		for _, r := range part[2:] {
			if !strings.ContainsRune("0123456789abcdefABCDEF", r) {
				return false
			}
		}

		return true
	}

	for _, r := range part {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}

func parseIPv4Part(part string) (uint64, bool) {
	if part == "" {
		return 0, false
	}

	base := 10

	switch {
	case strings.HasPrefix(part, "0x") || strings.HasPrefix(part, "0X"):
		part = part[2:]
		base = 16

		if part == "" {
			return 0, true
		}

	case len(part) > 1 && part[0] == '0':
		part = part[1:]
		base = 8
	}

	value, err := strconv.ParseUint(part, base, 32)
	if err != nil {
		return 0, false
	}

	return value, true
}

// publicURL resolves a reference like resolveURL and also requires the
// result to point at the public internet. Credentials in the authority and
// backslashes are refused outright: browsers and Go split such URLs into host
// and path differently, so the host checked here might not be the one loaded.
func publicURL(base *url.URL, raw string) string {
	if strings.ContainsRune(raw, '\\') {
		return ""
	}

	resolved := resolveURL(base, raw)
	if resolved == "" {
		return ""
	}

	parsed, err := url.Parse(resolved)
	if err != nil || parsed.User != nil || !publicHost(parsed.Hostname()) {
		return ""
	}

	return resolved
}

// hostPort is the host and port a request to the URL dials, the way
// http.Transport spells it.
func hostPort(u *url.URL) string {
	port := u.Port()

	if port == "" {
		port = "80"

		if u.Scheme == "https" {
			port = "443"
		}
	}

	return net.JoinHostPort(strings.ToLower(u.Hostname()), port)
}

type trustedHostKey struct{}

// withTrustedHost marks the configured feed's own host as allowed to be
// internal. A feed served from the local network is an administrator's
// choice; somewhere a feed redirects to is not.
func withTrustedHost(ctx context.Context, feedURL *url.URL) context.Context {
	return context.WithValue(ctx, trustedHostKey{}, hostPort(feedURL))
}

// checkRedirect refuses redirects that leave http(s), carry credentials, loop
// too long, or name an internal host. A redirect within the configured feed's
// own host is allowed whatever that host is.
func checkRedirect(request *http.Request, via []*http.Request) error {
	if len(via) >= maxRedirects {
		return fmt.Errorf("stopped after %d redirects", maxRedirects)
	}

	target := request.URL

	if target.Scheme != "http" && target.Scheme != "https" {
		return fmt.Errorf("redirect to a %q URL", target.Scheme)
	}

	if target.User != nil {
		return errors.New("redirect to a URL with credentials")
	}

	if hostPort(target) == hostPort(via[0].URL) {
		return nil
	}

	if !publicHost(target.Hostname()) {
		return fmt.Errorf(
			"%w: redirect to %s",
			ErrInternalAddress,
			target.Hostname(),
		)
	}

	return nil
}

// guardedDialer connects to the configured feed host as asked, and to any
// other host only after resolving it and checking every address it resolves
// to. It then dials the checked address itself, so DNS answering differently
// a moment later (rebinding) cannot slip an internal address in between.
type guardedDialer struct {
	dialer   *net.Dialer
	resolver func() Resolver
}

func (d *guardedDialer) DialContext(
	ctx context.Context,
	network string,
	address string,
) (net.Conn, error) {
	if trusted, _ := ctx.Value(trustedHostKey{}).(string); strings.EqualFold(trusted, address) {
		return d.dialer.DialContext(ctx, network, address)
	}

	host, portText, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	port, err := strconv.ParseUint(portText, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("invalid port %q", portText)
	}

	if !publicHost(host) {
		return nil, fmt.Errorf("%w: %s", ErrInternalAddress, host)
	}

	addrs, err := d.resolver().LookupNetIP(ctx, "ip", host)
	if err != nil {
		return nil, err
	}

	if len(addrs) == 0 {
		return nil, fmt.Errorf("no address for %s", host)
	}

	// Any internal answer disqualifies the host: an attacker controlling the
	// zone can mix one in among public ones.
	for _, addr := range addrs {
		if internalAddr(addr) {
			return nil, fmt.Errorf("%w: %s resolves to %s", ErrInternalAddress, host, addr)
		}
	}

	var lastErr error

	for _, addr := range addrs {
		conn, err := d.dialer.DialContext(
			ctx,
			network,
			netip.AddrPortFrom(addr.Unmap(), uint16(port)).String(),
		)
		if err == nil {
			return conn, nil
		}

		lastErr = err
	}

	return nil, lastErr
}

// newGuardedClient builds the feed client. The proxy from the environment is
// deliberately not used: the guard would check the proxy's address instead of
// the destination's.
func newGuardedClient(
	timeout time.Duration,
	resolver func() Resolver,
) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	transport.Proxy = nil
	transport.DialContext = (&guardedDialer{
		dialer: &net.Dialer{
			Timeout:   timeout,
			KeepAlive: 30 * time.Second,
		},
		resolver: resolver,
	}).DialContext

	return &http.Client{
		Timeout:       timeout,
		Transport:     transport,
		CheckRedirect: checkRedirect,
	}
}
