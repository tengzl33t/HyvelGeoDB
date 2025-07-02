/*
This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.

SPDX-License-Identifier: MPL-2.0

File: country.go
Description: Country DB specific functions
Author: tengzl33t
*/

package internal

import (
	"archive/zip"
	"encoding/csv"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func GetCountryCSVData(countryDBPath string, locations bool) (*zip.ReadCloser, []*zip.File) {
	archive, err := zip.OpenReader(countryDBPath)
	if err != nil {
		panic(err)
	}
	var DBs []*zip.File

	for _, archivedFile := range archive.File {
		if strings.Contains(archivedFile.Name, "GeoLite2-Country-Locations") && locations {
			DBs = append(DBs, archivedFile)
		}
		if strings.Contains(archivedFile.Name, "GeoLite2-Country-Blocks") && !locations {
			DBs = append(DBs, archivedFile)
		}
	}
	return archive, DBs
}

func PrepareCountryCSVData(countryDBPath string) map[*net.IPNet]int {
	IPRanges := make(map[*net.IPNet]int)
	archive, countryDBs := GetCountryCSVData(countryDBPath, false)
	defer archive.Close()

	for _, countryPath := range countryDBs {
		countryDB, err := countryPath.Open()
		if err != nil {
			panic(err)
		}
		csvReader := csv.NewReader(countryDB)
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

			geoNameID, err := strconv.Atoi(record[1])
			if err != nil {
				geoNameID, err = strconv.Atoi(record[2])
				if err != nil {
					panic(err)
				}
			}
			IPRanges[prefix] = geoNameID

		}
		countryDB.Close()
	}
	return IPRanges
}

func PrepareDBPaths(DBDir string) (map[string]DBStruct, map[string]DBStruct) {
	DBInits := GetDBStructs()

	countryDBs := make(map[string]DBStruct)
	ASNDBs := make(map[string]DBStruct)

	fullDirPath, _ := filepath.Abs(DBDir)
	dirEntries, err := os.ReadDir(DBDir)
	if err != nil {
		panic(err)
	}

	for _, entry := range dirEntries {
		for k, v := range DBInits {
			if strings.Contains(entry.Name(), k) {
				v.Path = filepath.Join(fullDirPath, entry.Name())
				if strings.Contains(entry.Name(), "ASN") {
					ASNDBs[k] = v
				} else if strings.Contains(entry.Name(), "Country") {
					countryDBs[k] = v
				} else if strings.Contains(entry.Name(), "IPFire") || strings.Contains(entry.Name(), "IPInfo") {
					ASNDBs[k] = v
					countryDBs[k] = v
				}
				break
			}
		}
	}
	return countryDBs, ASNDBs
}
