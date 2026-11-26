package providers

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// HostProxy is a localhost HTTP proxy (CONNECT for TLS, absolute-URL
// forwarding for plain HTTP) that admits only the listed hosts. A
// sandboxed run reaches it through HTTPS_PROXY while the seatbelt profile
// admits only its port, which is how "network only to the provider's API
// hosts" is enforced when seatbelt cannot name hosts (package doc).
type HostProxy struct {
	ln      net.Listener
	srv     *http.Server
	hosts   map[string]bool
	denied  atomic.Int64
	closeMu sync.Once
}

// StartHostProxy listens on 127.0.0.1 and serves until Close. A host entry
// admits itself and its subdomains.
func StartHostProxy(hosts []string) (*HostProxy, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	p := &HostProxy{ln: ln, hosts: map[string]bool{}}
	for _, h := range hosts {
		h = strings.ToLower(strings.TrimSpace(h))
		if h != "" {
			p.hosts[h] = true
		}
	}
	p.srv = &http.Server{Handler: p, ReadHeaderTimeout: 30 * time.Second}
	go p.srv.Serve(ln)
	return p, nil
}

// Addr is host:port.
func (p *HostProxy) Addr() string { return p.ln.Addr().String() }

// Port is the listening port.
func (p *HostProxy) Port() int { return p.ln.Addr().(*net.TCPAddr).Port }

// URL is the proxy URL for the HTTPS_PROXY variables.
func (p *HostProxy) URL() string { return "http://" + p.Addr() }

// DeniedCount is how many requests were refused.
func (p *HostProxy) DeniedCount() int { return int(p.denied.Load()) }

// Close stops the listener.
func (p *HostProxy) Close() error {
	var err error
	p.closeMu.Do(func() { err = p.srv.Close() })
	return err
}

// Allowed reports whether host (without port) is admitted.
func (p *HostProxy) Allowed(host string) bool {
	host = strings.ToLower(host)
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	for h := range p.hosts {
		if host == h || strings.HasSuffix(host, "."+h) {
			return true
		}
	}
	return false
}

// ServeHTTP implements http.Handler.
func (p *HostProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	if r.Method != http.MethodConnect && r.URL.Host != "" {
		host = r.URL.Host
	}
	if !p.Allowed(host) {
		p.denied.Add(1)
		http.Error(w, "bookwriter sandbox: host not allowed: "+host, http.StatusForbidden)
		return
	}
	if r.Method == http.MethodConnect {
		p.tunnel(w, r, host)
		return
	}
	if r.URL.Host == "" {
		http.Error(w, "bookwriter sandbox: proxy requests only", http.StatusBadRequest)
		return
	}
	out := r.Clone(r.Context())
	out.RequestURI = ""
	out.Header.Del("Proxy-Connection")
	resp, err := proxyTransport.RoundTrip(out)
	if err != nil {
		http.Error(w, "bookwriter sandbox: upstream: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

var proxyTransport = &http.Transport{Proxy: nil, DialContext: (&net.Dialer{Timeout: 20 * time.Second}).DialContext}

func (p *HostProxy) tunnel(w http.ResponseWriter, r *http.Request, host string) {
	if _, _, err := net.SplitHostPort(host); err != nil {
		host = net.JoinHostPort(host, "443")
	}
	upstream, err := net.DialTimeout("tcp", host, 20*time.Second)
	if err != nil {
		http.Error(w, "bookwriter sandbox: dial: "+err.Error(), http.StatusBadGateway)
		return
	}
	hj, ok := w.(http.Hijacker)
	if !ok {
		upstream.Close()
		http.Error(w, "bookwriter sandbox: hijack unsupported", http.StatusInternalServerError)
		return
	}
	client, buf, err := hj.Hijack()
	if err != nil {
		upstream.Close()
		return
	}
	fmt.Fprintf(buf, "HTTP/1.1 200 Connection Established\r\n\r\n")
	buf.Flush()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); io.Copy(upstream, buf); closeWrite(upstream) }()
	go func() { defer wg.Done(); io.Copy(client, upstream); closeWrite(client) }()
	wg.Wait()
	upstream.Close()
	client.Close()
}

func closeWrite(c net.Conn) {
	if t, ok := c.(*net.TCPConn); ok {
		t.CloseWrite()
	}
}

// proxyEnv are the variables that point a child at the proxy.
func proxyEnv(url string) map[string]string {
	return map[string]string{
		"HTTPS_PROXY": url, "HTTP_PROXY": url, "https_proxy": url, "http_proxy": url, "ALL_PROXY": url,
	}
}
