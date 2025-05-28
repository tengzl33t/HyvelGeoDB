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
