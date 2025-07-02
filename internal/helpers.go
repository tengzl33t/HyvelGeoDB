/*
This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.

SPDX-License-Identifier: MPL-2.0

File: helpers.go
Description: Internal functions
Author: tengzl33t
*/

package internal

func GetDBStructs() map[string]DBStruct {
	var emptyInterface interface{}

	return map[string]DBStruct{
		"GeoLite2": {
			"GeoLite2",
			"",
			4,
			"CSV",
			emptyInterface,
			[]string{},
			[]string{},
		},
		"IPFire": {
			"IPFire",
			"",
			4,
			"BIN",
			emptyInterface,
			[]string{},
			[]string{},
		},
		"IPInfo": {
			"IPInfo",
			"",
			4,
			"MMDB",
			emptyInterface,
			[]string{"country_code"},
			[]string{"asn"},
		},
		"IPLocate": {
			"IPLocate",
			"",
			3,
			"MMDB",
			emptyInterface,
			[]string{"country_code"},
			[]string{"asn"},
		},
		"IP2Location": {
			"IP2Location",
			"",
			2,
			"MMDB",
			emptyInterface,
			[]string{"country", "iso_code"},
			[]string{"autonomous_system_number"},
		},
		"DBIP": {
			"DBIP",
			"",
			2,
			"MMDB",
			emptyInterface,
			[]string{"country", "iso_code"},
			[]string{"autonomous_system_number"},
		},
	}
}

func GetContinentCodes() []string {
	return []string{"AF", "AN", "AS", "EU", "NA", "OC", "SA"}
}

func GetMaxInCountryMap(input map[string]int) string {
	var maxValue int
	var result string
	for k, v := range input {
		if v >= maxValue {
			maxValue = v
			result = k
		}
	}
	return result
}

func GetMaxInASNMap(input map[int]int) int {
	var maxValue int
	var result int
	for k, v := range input {
		if v >= maxValue {
			maxValue = v
			result = k
		}
	}
	return result
}
