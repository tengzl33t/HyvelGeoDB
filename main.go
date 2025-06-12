/*
This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.

SPDX-License-Identifier: MPL-2.0

File: main.go
Description: Main entrypoint
Author: tengzl33t
*/

package main

import (
	"HyvelGeoDB/internal"
	"archive/zip"
	"compress/gzip"
	"encoding/csv"
	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
	"github.com/oschwald/maxminddb-golang"
	"gitlab.com/yawning/gibloc"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

func getLangLocations(langPaths []*zip.File) (map[int]mmdbtype.Map, map[string]int) {
	countries := make(map[int]mmdbtype.Map)
	countriesAssignMap := make(map[string]int)

	for _, langPath := range langPaths {
		langDB, err := langPath.Open()
		if err != nil {
			log.Fatal(err)
		}
		csvReader := csv.NewReader(langDB)
		_, _ = csvReader.Read()

		for {
			record, err := csvReader.Read()
			if err == io.EOF {
				break
			}
			gotGeoNameID, err := strconv.Atoi(record[0])

			if _, ok := countries[gotGeoNameID]; !ok {
				countries[gotGeoNameID] = mmdbtype.Map{
					"geoname_id": mmdbtype.Uint32(gotGeoNameID),
					"iso_code":   mmdbtype.String(record[4]),
					"names": mmdbtype.Map{
						mmdbtype.String(record[1]): mmdbtype.String(record[5]),
					},
				}
				countriesAssignMap[record[4]] = gotGeoNameID
			} else {
				countries[gotGeoNameID]["names"].(mmdbtype.Map)[mmdbtype.String(record[1])] = mmdbtype.String(record[5])
			}

		}
		langDB.Close()
	}

	return countries, countriesAssignMap
}

func GetDBReadStreams(DBs map[string]internal.DBStruct) map[string]internal.DBStruct {
	for DBName, DB := range DBs {
		if DB.Type == "CSV" {
			continue
		}
		fi, err := os.Open(DB.Path)
		if err != nil {
			panic(err)
		}
		fz, err := gzip.NewReader(fi)
		if err != nil {
			panic(err)
		}
		stream, err := io.ReadAll(fz)
		if err != nil {
			panic(err)
		}
		if DB.Type == "MMDB" {
			var DBReader *maxminddb.Reader
			DBReader, err = maxminddb.FromBytes(stream)
			if err != nil {
				log.Fatal(err)
			}
			DB.Reader = DBReader
			DBs[DBName] = DB
		}
		if DB.Type == "BIN" && DB.Name == "IPFire" {
			var DBReader *gibloc.DB
			DBReader, err := gibloc.New(stream)
			if err != nil {
				panic(err)
			}
			DB.Reader = DBReader
			DBs[DBName] = DB
		}
		fi.Close()
		fz.Close()
	}
	return DBs
}

func collectDBASNs(ASNDBs map[string]internal.DBStruct, outputDirPath string) {
	CSVIPRanges, ASNNameMap := internal.PrepareASNCSVData(ASNDBs["GeoLite2"].Path)

	writer, err := mmdbwriter.New(mmdbwriter.Options{
		DatabaseType: "GeoLite2",
	})
	if err != nil {
		panic(err.Error())
	}

	for key, value := range CSVIPRanges {
		ASNResults := make(map[int]int)
		ASNResults[value] = ASNDBs["GeoLite2"].Priority

		for DBName, DB := range ASNDBs {
			if DBName == "GeoLite2" {
				continue
			}
			var ASNResult int

			if DB.Type == "MMDB" {
				reader := DB.Reader.(*maxminddb.Reader)
				var gotData interface{}
				_, ok, err := reader.LookupNetwork(key.IP, &gotData)
				if err != nil {
					panic(err.Error())
				}
				if !ok {
					continue
				}
				for _, path := range DB.ASNSearchPaths {
					gotData = gotData.(map[string]interface{})[path]
				}
				switch gotData.(type) {
				case int:
					ASNResult = gotData.(int)
				case uint32:
					ASNResult = int(gotData.(uint32))
				case uint64:
					ASNResult = int(gotData.(uint64))
				case string:
					res, err := strconv.Atoi(strings.ReplaceAll(gotData.(string), "AS", ""))
					if err != nil {
						panic(err)
					}
					ASNResult = res
				case nil:
					continue
				default:
					panic("Got unprocessable type.")
				}
			}
			if DB.Type == "BIN" && DB.Name == "IPFire" {
				reader := DB.Reader.(*gibloc.DB)
				entry := reader.QueryIP(key.IP)
				if entry != nil {
					ASNResult = int(entry.ASN)
				}
			}
			if ASNResult == 0 {
				continue
			}
			if val, ok := ASNResults[ASNResult]; ok {
				ASNResults[ASNResult] = val + DB.Priority
			} else {
				ASNResults[ASNResult] = DB.Priority
			}
		}
		gotASN := internal.GetMaxInASNMap(ASNResults)

		err = writer.Insert(
			key,
			mmdbtype.Map{
				"autonomous_system_number":       mmdbtype.Uint32(gotASN),
				"autonomous_system_organization": mmdbtype.String(ASNNameMap[gotASN]),
			},
		)
		if err != nil {
			panic(err)
		}

	}
	mmdbOutputFile, err := os.Create(filepath.Join(outputDirPath, "HyvelGeoDB-ASN.mmdb"))
	if err != nil {
		panic(err)
	}

	_, err = writer.WriteTo(mmdbOutputFile)
	if err != nil {
		panic(err)
	}
}

func collectDBCountries(countryDBs map[string]internal.DBStruct, outputDirPath string) {
	CSVIPRanges := internal.PrepareCountryCSVData(countryDBs["GeoLite2"].Path)
	archive, langDBs := internal.GetCountryCSVData(countryDBs["GeoLite2"].Path, true)
	defer archive.Close()

	countries, countryMap := getLangLocations(langDBs)

	writer, err := mmdbwriter.New(mmdbwriter.Options{
		DatabaseType: "GeoLite2",
	})
	if err != nil {
		panic(err)
	}

	for key, value := range CSVIPRanges {
		locationResults := make(map[string]int)
		gotISOCode := string(countries[value]["iso_code"].(mmdbtype.String))

		locationResults[gotISOCode] = countryDBs["GeoLite2"].Priority

		for DBName, DB := range countryDBs {
			if DBName == "GeoLite2" {
				continue
			}
			var locationResult string
			if DB.Type == "MMDB" {
				reader := DB.Reader.(*maxminddb.Reader)
				var gotData interface{}
				_, ok, err := reader.LookupNetwork(key.IP, &gotData)
				if err != nil {
					panic(err.Error())
				}
				if !ok {
					continue
				}
				for _, path := range DB.CountrySearchPaths {
					gotData = gotData.(map[string]interface{})[path]
				}
				locationResult = gotData.(string)
			}
			if DBName == "IPFire" {
				reader := DB.Reader.(*gibloc.DB)
				entry := reader.QueryIP(key.IP)
				if entry != nil {
					locationResult = entry.CountryCode
				}
			}
			if locationResult == "" || slices.Contains(internal.GetContinentCodes(), locationResult) {
				continue
			}
			if val, ok := locationResults[locationResult]; ok {
				locationResults[locationResult] = val + DB.Priority
			} else {
				locationResults[locationResult] = DB.Priority
			}
		}

		gotCountry := internal.GetMaxInCountryMap(locationResults)
		countryGeoID := countryMap[gotCountry]

		err = writer.Insert(key, mmdbtype.Map{"country": countries[countryGeoID]})
		if err != nil {
			panic(err)
		}
	}

	mmdbOutputFile, err := os.Create(filepath.Join(outputDirPath, "HyvelGeoDB-Country.mmdb"))
	if err != nil {
		panic(err)
	}

	_, err = writer.WriteTo(mmdbOutputFile)
	if err != nil {
		panic(err)
	}
}

func createMMDBs(sourcesDirPath string, outputDirPath string) {
	countryDBs, asnDBs := internal.PrepareDBPaths(sourcesDirPath)
	countryDBStreams := GetDBReadStreams(countryDBs)
	collectDBCountries(countryDBStreams, outputDirPath)

	ASNDBStreams := GetDBReadStreams(asnDBs)
	collectDBASNs(ASNDBStreams, outputDirPath)
}

func main() {
	cmdArgs := os.Args[1:]
	if len(cmdArgs) < 2 {
		println("usage: main <sources directory> <output directory>")
		os.Exit(1)
	}
	createMMDBs(cmdArgs[0], cmdArgs[1])
}
