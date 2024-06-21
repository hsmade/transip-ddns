package ddns

import (
	"fmt"
	"github.com/transip/gotransip/v6/domain"
	"github.com/transip/gotransip/v6/repository"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

type DDNS struct {
	Client      repository.Client
	DomainNames []string
}

// Update will go through all domain names and update them with the current IP, if needed
func (d *DDNS) Update() error {
	ip, err := d.getIP()
	if err != nil {
		return fmt.Errorf("updating: %w", err)
	}
	slog.Info("found public IP", "ip", ip)

	for _, domainName := range d.DomainNames {
		slog.Info("checking domainName", "domainName", domainName)
		entry, zone, err := d.getDnsEntry(domainName)
		if err != nil {
			return fmt.Errorf("fetching ip from domainName '%s': %w", domainName, err)
		}

		if entry.Content == ip {
			slog.Info("domainName is up to date", "domainName", domainName)
			continue
		}

		slog.Info("updating domainName", "host", entry.Name, "zone", zone)
		entry.Content = ip
		err = d.updateRecord(zone, entry)
		if err != nil {
			return fmt.Errorf("updating domainName: %w", err)
		}
	}

	return nil
}

// updateRecord will store a DNS Entry in the domain
func (d *DDNS) updateRecord(zone string, entry *domain.DNSEntry) error {
	domainRepo := domain.Repository{Client: d.Client}
	slog.Warn("Updating zone", "zone", zone, "entry", entry)
	err := domainRepo.UpdateDNSEntry(zone, *entry)
	if err != nil {
		return fmt.Errorf("updating zone '%s': %w", zone, err)
	}
	return nil
}

// getDnsEntry finds the DNS Entry and zone name for a domainName
func (d *DDNS) getDnsEntry(domainName string) (*domain.DNSEntry, string, error) {
	domainObject, err := d.getZone(domainName)
	if err != nil {
		return nil, "", fmt.Errorf("finding domain for domainName '%s': %w", domainName, err)
	}
	zone := domainObject.Name
	host := strings.TrimSuffix(domainName, "."+zone)
	slog.Debug("found domain", "host", host, "zone", zone)

	domainRepo := domain.Repository{Client: d.Client}
	entries, err := domainRepo.GetDNSEntries(domainObject.Name)
	for _, entry := range entries {
		slog.Debug("checking entry", "host", host, "zone", zone, "entry", entry)
		if entry.Type != "A" {
			continue
		}
		if entry.Name == host {
			return &entry, zone, nil
		}
	}
	return nil, "", fmt.Errorf("finding domainName '%s': not found", domainName)
}

// getZone finds the Zone for a domainName
func (d *DDNS) getZone(domainName string) (*domain.Domain, error) {
	domainNameParts := strings.Split(domainName, ".")

	domainRepo := domain.Repository{Client: d.Client}
	for i := 0; i < len(domainNameParts); i++ {
		slog.Debug("trying zone", "domainName", domainName, "zone", domainNameParts[i:])
		zone := strings.Join(domainNameParts[i:], ".")
		domainObject, err := domainRepo.GetByDomainName(zone)
		if err != nil || domainObject.Name == "" {
			continue
		}
		slog.Debug("found zone", "domainName", domainName, "zone", domainObject.Name)
		return &domainObject, nil
	}
	return nil, fmt.Errorf("no zone found for '%s'", domainName)
}

// getIP finds the current public IP
func (d *DDNS) getIP() (string, error) {
	res, err := http.Get("https://ifconfig.me")
	if err != nil {
		return "", fmt.Errorf("getting ip from ifconfig.me: %w", err)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}
	return string(data), nil
}
