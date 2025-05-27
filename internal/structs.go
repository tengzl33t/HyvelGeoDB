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

type DBStruct struct {
	Name               string
	Path               string
	Priority           int
	Type               string
	Reader             interface{}
	CountrySearchPaths []string
	ASNSearchPaths     []string
}
