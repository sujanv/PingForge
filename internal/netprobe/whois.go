package netprobe

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// WHOISResult contains domain or IP registration and registrar information.
type WHOISResult struct {
	Query        string   `json:"query"`
	Registrar    string   `json:"registrar,omitempty"`
	CreatedDate  string   `json:"created_date,omitempty"`
	ExpiryDate   string   `json:"expiry_date,omitempty"`
	UpdatedDate  string   `json:"updated_date,omitempty"`
	NameServers  []string `json:"nameservers,omitempty"`
	Status       []string `json:"status,omitempty"`
	AbuseContact string   `json:"abuse_contact,omitempty"`
	Source       string   `json:"source"`
	Raw          string   `json:"raw,omitempty"`
	Error        string   `json:"error,omitempty"`
}

// QueryWHOIS performs an RFC 3912 TCP WHOIS query with fallback to RDAP JSON API.
func QueryWHOIS(ctx context.Context, target string, timeout time.Duration) *WHOISResult {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	res := &WHOISResult{
		Query:  target,
		Source: "WHOIS-TCP",
	}

	// Determine whois server
	whoisServer := "whois.iana.org:43"
	if strings.Contains(target, ".") && !strings.ContainsAny(target, ":") {
		parts := strings.Split(target, ".")
		tld := parts[len(parts)-1]
		if tld == "com" || tld == "net" {
			whoisServer = "whois.verisign-grs.com:43"
		}
	}

	// Establish dialer with explicit timeout and TCP keep-alive
	d := net.Dialer{Timeout: timeout, KeepAlive: 30 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", whoisServer)
	if err != nil {
		// Fallback to RDAP
		return queryRDAP(ctx, target, timeout)
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(timeout))
	_, err = fmt.Fprintf(conn, "%s\r\n", target)
	if err != nil {
		return queryRDAP(ctx, target, timeout)
	}

	var sb strings.Builder
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		sb.WriteString(line + "\n")

		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "registrar:") && res.Registrar == "" {
			res.Registrar = strings.TrimSpace(line[strings.Index(line, ":")+1:])
		} else if (strings.Contains(lower, "creation date:") || strings.Contains(lower, "created:")) && res.CreatedDate == "" {
			res.CreatedDate = strings.TrimSpace(line[strings.Index(line, ":")+1:])
		} else if (strings.Contains(lower, "registry expiry date:") || strings.Contains(lower, "expiration date:")) && res.ExpiryDate == "" {
			res.ExpiryDate = strings.TrimSpace(line[strings.Index(line, ":")+1:])
		} else if strings.Contains(lower, "name server:") {
			ns := strings.TrimSpace(line[strings.Index(line, ":")+1:])
			if ns != "" {
				res.NameServers = append(res.NameServers, strings.ToLower(ns))
			}
		}
	}

	res.Raw = sb.String()
	if res.Registrar == "" && res.CreatedDate == "" {
		return queryRDAP(ctx, target, timeout)
	}

	return res
}

func queryRDAP(ctx context.Context, target string, timeout time.Duration) *WHOISResult {
	res := &WHOISResult{
		Query:  target,
		Source: "RDAP-HTTP",
	}

	apiURL := fmt.Sprintf("https://rdap.org/domain/%s", target)
	if net.ParseIP(target) != nil {
		apiURL = fmt.Sprintf("https://rdap.org/ip/%s", target)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	req.Header.Set("Accept", "application/rdap+json, application/json")

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		res.Error = fmt.Sprintf("rdap query failed: %v", err)
		return res
	}
	defer resp.Body.Close()

	var payload struct {
		Handle string `json:"handle"`
		Events []struct {
			Action string `json:"eventAction"`
			Date   string `json:"eventDate"`
		} `json:"events"`
		Nameservers []struct {
			LdhName string `json:"ldhName"`
		} `json:"nameservers"`
		Status []string `json:"status"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		res.Error = err.Error()
		return res
	}

	res.Registrar = payload.Handle
	res.Status = payload.Status
	for _, ns := range payload.Nameservers {
		res.NameServers = append(res.NameServers, ns.LdhName)
	}
	for _, ev := range payload.Events {
		if ev.Action == "registration" {
			res.CreatedDate = ev.Date
		} else if ev.Action == "expiration" {
			res.ExpiryDate = ev.Date
		} else if ev.Action == "last changed" {
			res.UpdatedDate = ev.Date
		}
	}

	return res
}
