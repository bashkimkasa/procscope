package platform

import "testing"

func TestExtractDNSQueriesFromCommandLine(t *testing.T) {
	t.Parallel()

	cmd := "C:/Program Files/Google/Chrome/Application/chrome.exe https://example.com/path --proxy https://api.github.com/v3"
	queries := extractDNSQueriesFromCommandLine(cmd, 1234)

	if len(queries) < 2 {
		t.Fatalf("expected at least 2 DNS queries, got %d", len(queries))
	}

	seen := map[string]bool{}
	for _, q := range queries {
		seen[q.Query] = true
	}

	if !seen["example.com"] {
		t.Fatalf("expected example.com in extracted queries: %#v", queries)
	}
	if !seen["api.github.com"] {
		t.Fatalf("expected api.github.com in extracted queries: %#v", queries)
	}
}

func TestExtractDNSQueriesIgnoresNonHosts(t *testing.T) {
	t.Parallel()

	cmd := "cmd /c echo hello 127.0.0.1 localhost"
	queries := extractDNSQueriesFromCommandLine(cmd, 1234)
	if len(queries) != 0 {
		t.Fatalf("expected no hostname queries, got %#v", queries)
	}
}

func TestExtractDNSQueriesFromHostnameFlags(t *testing.T) {
	t.Parallel()

	cmd := "app.exe --host=api.example.com --server api.github.com --endpoint=example.org:8443 --url=https://internal.example.net/path"
	queries := extractDNSQueriesFromCommandLine(cmd, 5678)

	if len(queries) < 4 {
		t.Fatalf("expected at least 4 hostname-like queries, got %d: %#v", len(queries), queries)
	}

	seen := map[string]bool{}
	for _, q := range queries {
		seen[q.Query] = true
	}

	for _, want := range []string{"api.example.com", "api.github.com", "example.org", "internal.example.net"} {
		if !seen[want] {
			t.Fatalf("expected %q in extracted queries: %#v", want, queries)
		}
	}
}
