package platform

import (
	"regexp"
	"strings"
	"time"

	"github.com/bashkimkasa/procscope/internal/models"
)

func extractDNSQueriesFromCommandLine(commandLine string, pid int) []*models.DNSQuery {
	queries := make([]*models.DNSQuery, 0)
	if strings.TrimSpace(commandLine) == "" {
		return queries
	}

	cmdLine := strings.ReplaceAll(commandLine, "\x00", " ")
	cmdLine = strings.ReplaceAll(cmdLine, "\\", "/")

	urlRe := regexp.MustCompile(`https?://([A-Za-z0-9.-]+)`)
	matches := urlRe.FindAllStringSubmatch(cmdLine, -1)
	seen := make(map[string]bool)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		host := normalizeHostname(match[1])
		if host == "" || !seen[host] {
			if host != "" {
				seen[host] = true
				queries = append(queries, newDNSQuery(host, pid))
			}
		}
	}

	flagPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?:--|/)(?:host|server|hostname|domain|url|endpoint|target|address|addr|proxy)(?:=|:)\s*([A-Za-z0-9.-]+)`),
		regexp.MustCompile(`(?:--|/)(?:host|server|hostname|domain|url|endpoint|target|address|addr|proxy)\s+([A-Za-z0-9.-]+)`),
	}
	for _, re := range flagPatterns {
		for _, match := range re.FindAllStringSubmatch(cmdLine, -1) {
			if len(match) < 2 {
				continue
			}
			host := normalizeHostname(match[1])
			if host == "" || seen[host] {
				continue
			}
			seen[host] = true
			queries = append(queries, newDNSQuery(host, pid))
		}
	}

	return queries
}

func normalizeHostname(value string) string {
	host := strings.TrimSpace(strings.ToLower(value))
	host = strings.Trim(host, "\"'[](){}<> ")
	host = strings.TrimSuffix(host, "/")
	host = strings.TrimSuffix(host, ":443")
	host = strings.TrimSuffix(host, ":80")
	if host == "" || strings.Contains(host, " ") || strings.Contains(host, "\\") || strings.Contains(host, ":") {
		if strings.Contains(host, ":") && !strings.Contains(host, ".") {
			return ""
		}
	}
	if host == "localhost" || host == "127.0.0.1" || strings.HasPrefix(host, "127.") || strings.HasPrefix(host, "0.0.0.0") || strings.HasPrefix(host, "::") {
		return ""
	}
	if regexp.MustCompile(`^\d+(?:\.\d+){3}$`).MatchString(host) {
		return ""
	}
	if strings.Count(host, ".") == 0 {
		return ""
	}
	return host
}

func newDNSQuery(host string, pid int) *models.DNSQuery {
	return &models.DNSQuery{
		Query:     host,
		QueryType: "A",
		Response:  "",
		Timestamp: time.Now(),
		PID:       pid,
	}
}
