package geoip

import (
	"fmt"
	"net"

	"github.com/oschwald/maxminddb-golang"
)

// Record holds the GeoIP data we extract per lookup.
type Record struct {
	Country struct {
		ISOCode string            `maxminddb:"iso_code"`
		Names   map[string]string `maxminddb:"names"`
	} `maxminddb:"country"`
	City struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"city"`
}

// Reader wraps a MaxMind database for IP → geography lookups.
type Reader struct {
	db *maxminddb.Reader
}

// Open opens the MaxMind database file at path.
func Open(path string) (*Reader, error) {
	db, err := maxminddb.Open(path)
	if err != nil {
		return nil, fmt.Errorf("geoip: open %s: %w", path, err)
	}
	return &Reader{db: db}, nil
}

// Lookup returns geographic information for ip.
// Returns an empty Record (not an error) when the IP is not found.
func (r *Reader) Lookup(ip net.IP) (Record, error) {
	var rec Record
	if err := r.db.Lookup(ip, &rec); err != nil {
		return Record{}, fmt.Errorf("geoip: lookup %s: %w", ip, err)
	}
	return rec, nil
}

// Close releases the database file handle.
func (r *Reader) Close() error {
	return r.db.Close()
}
