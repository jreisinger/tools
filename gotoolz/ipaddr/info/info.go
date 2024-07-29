// Package info provides information about IP addresses.
package info

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

type Info struct {
	DnsNames    []string `json:"dnsNames"`
	Org         string   `json:"org"`
	Geolocation `json:"geolocation"`
	AWS         `json:"aws"`
}

type Geolocation struct {
	City    string `json:"city"`
	Country string `json:"country"`
}

func Get(ip net.IP) (Info, error) {
	var info Info

	dnsNames, err := getDnsNames(ip)
	if err != nil {
		return info, err
	}
	info.DnsNames = dnsNames

	ii, err := getIpInfo(ip)
	if err != nil {
		return info, err
	}
	info.Org = ii.Org
	info.City = ii.City
	info.Country = ii.Country

	aws, err := isOnAWS(ip)
	if err != nil {
		return info, err
	}
	info.AWS = aws

	return info, err
}

func getDnsNames(ip net.IP) ([]string, error) {
	names, err := net.LookupAddr(ip.String())
	if err != nil {
		if len(names) == 0 {
			// IP address does not resolve to
			// any names, ignore this error.
		} else {
			return names, err
		}
	}
	for i := range names {
		names[i] = strings.TrimSuffix(names[i], ".")
	}
	return names, nil
}

type ipinfo struct {
	Hostname string `json:"hostname"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Org      string `json:"org"`
	Timezone string `json:"timezone"`
}

func getIpInfo(ip net.IP) (ipinfo, error) {
	var i ipinfo
	resp, err := http.Get("https://ipinfo.io/" + ip.String())
	if err != nil {
		return i, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return i, err
	}
	if err := json.Unmarshal(b, &i); err != nil {
		return i, err
	}
	return i, nil
}

type AWS struct {
	IsOn               bool     `json:"isOn"`
	IpPrefix           string   `json:"ipPrefix"`
	Region             string   `json:"region"`
	Services           []string `json:"services"`
	NetworkBorderGroup string   `json:"networkBorderGroup"`
}

func isOnAWS(ipaddr net.IP) (AWS, error) {
	var a AWS
	resp := struct {
		Prefixes []struct {
			IpPrefix           string `json:"ip_prefix"`
			Region             string `json:"region"`
			Service            string `json:"service"`
			NetworkBorderGroup string `json:"network_border_group"`
		} `json:"prefixes"`
	}{}

	r, err := http.Get("https://ip-ranges.amazonaws.com/ip-ranges.json")
	if err != nil {
		return a, err
	}
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&resp); err != nil {
		return a, err
	}

	for _, prefix := range resp.Prefixes {
		_, network, err := net.ParseCIDR(prefix.IpPrefix)
		if err != nil {
			return a, fmt.Errorf("parse CIDR %q: %v", prefix.IpPrefix, err)
		}
		if network.Contains(ipaddr) {
			a.IsOn = true
			a.IpPrefix = prefix.IpPrefix
			a.NetworkBorderGroup = prefix.NetworkBorderGroup
			a.Region = prefix.Region
			a.Services = append(a.Services, prefix.Service)
		}

	}
	return a, nil
}
