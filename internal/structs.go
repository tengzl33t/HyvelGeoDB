package internal

import (
	"net/netip"
)

type IPRangeASNStruct struct {
	Range netip.Prefix
	ASN   int
	Org   string
}

type ASNDBStruct struct {
	Name     string
	Priority int
	IPRanges map[string]IPRangeASNStruct
}

type IPRangeCountryStruct struct {
	Range       netip.Prefix
	CountryCode string
}

type LocationStruct struct {
	Code  string
	Names map[string]string
}

type CountryDBStruct struct {
	Name     string
	IPRanges map[string]IPRangeCountryStruct
}

type InDBStruct struct {
	Priority int
	Path     string
}

type MMDBCountryStruct struct {
	GeoNameID int               `json:"geoname_id"`
	ISOCode   string            `json:"iso_code"`
	Names     map[string]string `json:"names"`
}

type MMDBASNStruct struct {
	ASN int    `json:"autonomous_system_number"`
	Org string `json:"autonomous_system_organization"`
}

type DBStruct struct {
	Name               string
	Path               string
	Priority           int
	Type               string
	Reader             interface{}
	CountrySearchPaths []string
	ASNSearchPaths     []string
}

type LocationResultStruct struct {
	Location string
	Priority int
}
