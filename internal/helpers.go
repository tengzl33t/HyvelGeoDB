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
			[]string{},
		},
		"IPLocate": {
			"IPLocate",
			"",
			3,
			"MMDB",
			emptyInterface,
			[]string{"country_code"},
			[]string{},
		},
		"IP2Location": {
			"IP2Location",
			"",
			2,
			"MMDB",
			emptyInterface,
			[]string{"country", "iso_code"},
			[]string{},
		},
		"DBIP": {
			"DBIP",
			"",
			2,
			"MMDB",
			emptyInterface,
			[]string{"country", "iso_code"},
			[]string{},
		},
	}
}

func GetContinentCodes() []string {
	return []string{"AF", "AN", "AS", "EU", "NA", "OC", "SA"}
}
