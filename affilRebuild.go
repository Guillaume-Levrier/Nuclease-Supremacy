package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

var collection = make(map[string][]map[string]interface{})
var eu = "Austria Italy Belgium Latvia Bulgaria Lithuania Croatia Luxembourg Cyprus Malta Czech Republic Netherlands Denmark Poland Estonia Portugal Finland Romania France Slovakia Germany Slovenia Greece Spain Hungary Sweden Ireland United Kingdom"

type docProfile struct {
	EU    int
	CN    int
	US    int
	Other int
}

var profileMap = make(map[docProfile]int)

func readArray(dataArray *json.Decoder, enc *json.Encoder) {

	_, err := dataArray.Token()
	if err != nil {
		log.Fatal(err)
	}

	for dataArray.More() {

		var m map[string]interface{}

		err := dataArray.Decode(&m)
		if err != nil {
			log.Fatal(err)
		}

		for k := range m {
			if k == "title-short" {

				mk := fmt.Sprintf("%v", m[k])

				affil := strings.TrimPrefix(mk, "{\"affiliations\":")

				if strings.Contains(affil, ",\"OA\":false}") {
					affil = strings.TrimSuffix(affil, ",\"OA\":false}")
				} else {
					affil = strings.TrimSuffix(affil, ",\"OA\":true}")
				}

				affils := strings.SplitAfter(affil, "},")

				thisProfile := docProfile{0, 0, 0, 0}

				for i := 0; i < len(affils); i++ {
					if strings.Contains(affils[i], "country") {

						start := strings.Index(affils[i], "ountry") + 9
						country := affils[i][start:]

						if strings.Contains(country, "\"") {

							country = country[:strings.Index(country, "\"")]

							if strings.Contains(eu, country) {
								collection["EU"] = append(collection["EU"], m)
								thisProfile.EU++
							} else {
								collection[country] = append(collection[country], m)
								switch country {
								case "China":
									thisProfile.CN++
								case "United States":
									thisProfile.US++
								default:
									thisProfile.Other++
								}
							}
							profileMap[thisProfile]++
						}
					}
				}
			}
		}

	}

	_, err = dataArray.Token()
	if err != nil {
		log.Fatal(err)
	}

}

func readFile(fileName string) {

	openFile, _ := os.Open(fileName)
	defer openFile.Close()

	data := io.Reader(openFile)
	writer := os.Stdout

	dec := json.NewDecoder(data)
	enc := json.NewEncoder(writer)
	readArray(dec, enc)

}

func sortFiles() {
	path, err := os.Getwd()
	files, err := ioutil.ReadDir(path)
	filerr := os.MkdirAll("sorted", 0755)

	if err != nil || filerr != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {

			readFile(file.Name())

		}
	}

}

func printCollection() {

	totalAmount := 0

	writeProfiles, _ := os.Create("profiles.json")
	defer writeProfiles.Close()

	writeProfiles.WriteString("[")
	for key, value := range profileMap {
		writeProfiles.WriteString("{\"EU\":" + strconv.Itoa(key.EU) + ",\"CN\":" + strconv.Itoa(key.CN) + ",\"US\":" + strconv.Itoa(key.US) + ",\"other\":" + strconv.Itoa(key.Other) + ",\"count\":" + strconv.Itoa(value) + "},")

	}
	writeProfiles.WriteString("]")

	writeReport, _ := os.Create("report.csv")
	defer writeReport.Close()

	writeReport.WriteString("country,amount\n")

	for key, value := range collection {

		writeReport.WriteString(key + "," + strconv.Itoa(len(value)) + "\n")

		writeCountry, _ := os.Create("sorted/" + key + ".json")

		defer writeCountry.Close()

		writeCountry.WriteString("[")
		for i, v := range value {
			totalAmount++
			if i > 0 {
				writeCountry.WriteString(",\n")
			}
			val, _ := json.Marshal(v)
			writeCountry.WriteString(string(val))

		}
		writeCountry.WriteString("]")
	}
	writeReport.WriteString("total," + strconv.Itoa(totalAmount))
}

func main() {
	sortFiles()
	printCollection()

}
