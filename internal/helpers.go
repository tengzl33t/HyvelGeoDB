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

func GetInterfaceSlice[P any](slice []P) []any {
	anys := make([]any, 0, len(slice))
	for _, v := range slice {
		anys = append(anys, v)
	}
	return anys
}

func GetContinentCodes() []string {
	return []string{"AF", "AN", "AS", "EU", "NA", "OC", "SA"}
}
