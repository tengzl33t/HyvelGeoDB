/*
This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.

SPDX-License-Identifier: MPL-2.0

File: structs.go
Description: Internal structs
Author: tengzl33t
*/

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
