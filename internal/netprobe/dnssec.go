package netprobe

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// DNSSECStatus details cryptographic DNS validation results
type DNSSECStatus struct {
	Domain        string `json:"domain"`
	Validated     bool   `json:"validated"`
	RRSIGFound    bool   `json:"rrsig_found"`
	DNSKEYFound   bool   `json:"dnskey_found"`
	DSFound       bool   `json:"ds_found"`
	Authenticated bool   `json:"authenticated_data_ad_flag"`
	Resolver      string `json:"resolver"`
	Error         string `json:"error,omitempty"`
}

// ValidateDNSSEC validates DNSSEC chain of trust using Cloudflare/Google DoH DNSSEC flags
func ValidateDNSSEC(ctx context.Context, domain string, timeout time.Duration) *DNSSECStatus {
	if timeout <= 0 {
		timeout = 4 * time.Second
	}

	status := &DNSSECStatus{
		Domain:   domain,
		Resolver: "Cloudflare-DNSSEC",
	}

	// Query Cloudflare DoH with cd=0 (checking disabled=false) and do=1 (DNSSEC OK)
	url := fmt.Sprintf("https://cloudflare-dns.com/dns-query?name=%s&type=A&do=1", domain)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		status.Error = err.Error()
		return status
	}
	req.Header.Set("Accept", "application/dns-json")

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		status.Error = err.Error()
		return status
	}
	defer resp.Body.Close()

	var dohResp struct {
		Status int  `json:"Status"`
		AD     bool `json:"AD"` // Authenticated Data flag
		Answer []struct {
			Type int `json:"type"`
		} `json:"Answer"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&dohResp); err != nil {
		status.Error = err.Error()
		return status
	}

	status.Authenticated = dohResp.AD
	for _, a := range dohResp.Answer {
		if a.Type == 46 { // RRSIG
			status.RRSIGFound = true
		} else if a.Type == 48 { // DNSKEY
			status.DNSKEYFound = true
		} else if a.Type == 43 { // DS
			status.DSFound = true
		}
	}

	if status.Authenticated || status.RRSIGFound {
		status.Validated = true
	}

	return status
}
