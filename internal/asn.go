/*
This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.

SPDX-License-Identifier: MPL-2.0

File: asn.go
Description: ASN DB specific functions
Author: tengzl33t
*/

package internal

import (
	"archive/zip"
	"encoding/csv"
	"io"
	"net"
	"strconv"
	"strings"
)

func PrepareASNCSVData(ASNDBPath string) (map[*net.IPNet]int, map[int]string) {
	IPRanges := make(map[*net.IPNet]int)
	ASNNameMap := make(map[int]string)

	archive, err := zip.OpenReader(ASNDBPath)
	if err != nil {
		panic(err)
	}
	defer archive.Close()
	var ASNDBs []*zip.File

	for _, archivedFile := range archive.File {
		if strings.Contains(archivedFile.Name, "GeoLite2-ASN-Blocks") {
			ASNDBs = append(ASNDBs, archivedFile)
		}
	}

	for _, ASNPath := range ASNDBs {
		ASNDB, err := ASNPath.Open()
		if err != nil {
			panic(err)
		}
		csvReader := csv.NewReader(ASNDB)

		_, _ = csvReader.Read()

		for {
			record, err := csvReader.Read()
			if err == io.EOF {
				break
			}

			_, prefix, err := net.ParseCIDR(record[0])
			if err != nil {
				panic(err)
			}

			asn, err := strconv.Atoi(record[1])
			if err != nil {
				panic(err)
			}

			IPRanges[prefix] = asn
			if _, ok := ASNNameMap[asn]; !ok {
				ASNNameMap[asn] = record[2]
			}
		}
		ASNDB.Close()
	}
	return IPRanges, ASNNameMap
}
