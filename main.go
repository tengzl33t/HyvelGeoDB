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
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

func getLangLocations(langPaths []*zip.File) (map[int]internal.LocationStruct, map[string]int) {
	countries := make(map[int]internal.LocationStruct)
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
				countries[gotGeoNameID] = internal.LocationStruct{
					Code: record[4],
					Names: map[string]string{
						record[1]: record[5],
					},
				}
			} else {
				countries[gotGeoNameID].Names[record[1]] = record[5]
			}
		}
		langDB.Close()
	}

	for key, value := range countries {
		countriesAssignMap[value.Code] = key
	}

	return countries, countriesAssignMap
}

func prepareASNMMDBData(ASNDBPath string) internal.ASNDBStruct {
	IPRanges := make(map[string]internal.IPRangeASNStruct)

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
		println(ASNPath.Name)
		ASNDB, err := ASNPath.Open()
		if err != nil {
			log.Fatal(err)
		}
		csvReader := csv.NewReader(ASNDB)

		_, _ = csvReader.Read()

		for {
			record, err := csvReader.Read()
			if err == io.EOF {
				break
			}

			prefix, err := netip.ParsePrefix(record[0])
			if err != nil {
				log.Fatal(err)
			}

			asn, err := strconv.Atoi(record[1])
			if err != nil {
				log.Fatal(err)
			}

			ipRangeStruct := internal.IPRangeASNStruct{
				Range: prefix,
				ASN:   asn,
				Org:   record[2],
			}
			IPRanges[record[0]] = ipRangeStruct

		}
		ASNDB.Close()
	}
	return internal.ASNDBStruct{
		Name:     "MMDB",
		Priority: 0,
		IPRanges: IPRanges,
	}
}

func getCountryCSVData(countryDBPath string, locations bool) (*zip.ReadCloser, []*zip.File) {
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

func prepareCountryCSVData(countryDBPath string) map[netip.Prefix]int {
	IPRanges := make(map[netip.Prefix]int)
	archive, countryDBs := getCountryCSVData(countryDBPath, false)
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

			prefix, err := netip.ParsePrefix(record[0])
			if err != nil {
				log.Fatal(err)
			}

			geoNameID, err := strconv.Atoi(record[1])
			if err != nil {
				geoNameID, err = strconv.Atoi(record[2])
				if err != nil {
					log.Fatal(err)
				}
			}
			IPRanges[prefix] = geoNameID

		}
		countryDB.Close()
	}
	return IPRanges
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

func getMaxInMap(input map[string]int) string {
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

func collectDBCountries(countryDBs map[string]internal.DBStruct) {
	CSVIPRanges := prepareCountryCSVData(countryDBs["GeoLite2"].Path)
	archive, langDBs := getCountryCSVData(countryDBs["GeoLite2"].Path, true)
	defer archive.Close()

	countries, countryMap := getLangLocations(langDBs)

	writer, err := mmdbwriter.New(mmdbwriter.Options{
		DatabaseType: "GeoLite2",
	})
	if err != nil {
		panic(err.Error())
	}

	var counter int

	for key, value := range CSVIPRanges {
		locationResults := make(map[string]int)
		locationResults[countries[value].Code] = countryDBs["GeoLite2"].Priority
		firstNetIP := net.ParseIP(key.Addr().String())
		counter++

		for _, DB := range countryDBs {
			var locationResult string

			if DB.Type == "MMDB" {
				reader := DB.Reader.(*maxminddb.Reader)
				var gotData interface{}
				_, ok, err := reader.LookupNetwork(firstNetIP, &gotData)
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
			if DB.Type == "BIN" && DB.Name == "IPFire" {
				reader := DB.Reader.(*gibloc.DB)
				entry := reader.QueryIP(firstNetIP)
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

		gotCountry := getMaxInMap(locationResults)
		countryGeoID := countryMap[gotCountry]

		newMap := mmdbtype.Map{}

		for k, v := range countries[countryGeoID].Names {
			newMap[mmdbtype.String(k)] = mmdbtype.String(v)
		}

		record := mmdbtype.Map{
			"country": mmdbtype.Map{
				"geoname_id": mmdbtype.Uint32(countryGeoID),
				"iso_code":   mmdbtype.String(gotCountry),
				"names":      newMap,
			},
		}

		_, ipnet, _ := net.ParseCIDR(key.String())
		err = writer.Insert(ipnet, record)
		if err != nil {
			log.Println(err)
		}
		println(key.String())
		for k, v := range locationResults {
			println(k, v)
		}

		if counter > 50 {
			break
		}
	}

	fh2, err := os.Create("country-scratch-out.mmdb")
	if err != nil {
		log.Fatal(err)
	}

	// write to the mmdb file
	_, err = writer.WriteTo(fh2)
	if err != nil {
		log.Fatal(err)
	}
}

func prepareDBPaths(DBDir string) (map[string]internal.DBStruct, map[string]internal.DBStruct) {
	DBInits := internal.GetDBStructs()

	countryDBs := make(map[string]internal.DBStruct)
	ASNDBs := make(map[string]internal.DBStruct)

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

func main() {
	countryDBs, _ := prepareDBPaths("sourcedbs/")
	countryDBs = GetDBReadStreams(countryDBs)
	collectDBCountries(countryDBs)
}
