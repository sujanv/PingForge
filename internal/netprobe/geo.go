package netprobe

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// IPGeoInfo represents IP Geolocation, ASN, and ISP details
type IPGeoInfo struct {
	IP          string  `json:"ip"`
	City        string  `json:"city"`
	Region      string  `json:"region"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	ASN         string  `json:"asn"`
	Org         string  `json:"org"`
	ISP         string  `json:"isp"`
	Timezone    string  `json:"timezone"`
}

// LookupIPGeo queries public or custom IP intelligence API
func LookupIPGeo(ctx context.Context, ip string, timeout time.Duration) (*IPGeoInfo, error) {
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	reqURL := fmt.Sprintf("https://ipapi.co/%s/json/", ip)
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "PingForge/3.0 Network-Diagnostics")

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ipapi returned status %d", resp.StatusCode)
	}

	var raw struct {
		IP        string  `json:"ip"`
		City      string  `json:"city"`
		Region    string  `json:"region"`
		Country   string  `json:"country_name"`
		CountryCode string `json:"country_code"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		ASN       string  `json:"asn"`
		Org       string  `json:"org"`
		Timezone  string  `json:"timezone"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	return &IPGeoInfo{
		IP:          raw.IP,
		City:        raw.City,
		Region:      raw.Region,
		Country:     raw.Country,
		CountryCode: raw.CountryCode,
		Latitude:    raw.Latitude,
		Longitude:   raw.Longitude,
		ASN:         raw.ASN,
		Org:         raw.Org,
		ISP:         raw.Org,
		Timezone:    raw.Timezone,
	}, nil
}
