package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"time"
)

var quoteMap = make(map[string]float64)

var journalProfileMap = make(map[string]interface{})

type doc struct {
	Issued          interface{}
	DOI             string
	Count           float64
	JournalKeywords interface{}
	JName           string
}

type categories struct {
	CN []doc
	EU []doc
	US []doc
}

var cat categories

type pandoDBfile struct {
	Name    string
	Id      string
	Date    string
	Content categories
}

func buildFile(area string) {

	openFile, _ := os.Open(area + ".json")
	defer openFile.Close()

	data := io.Reader(openFile)

	dataArray := json.NewDecoder(data)

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

		var newDoc doc

		for k := range m {
			switch k {
			case "issued":
				newDoc.Issued = m[k]
			case "DOI":
				DOI := m[k].(string)
				newDoc.DOI = DOI
				newDoc.Count = quoteMap[DOI]
			case "container-title":
				jourTitle := m[k].(string)
				newDoc.JournalKeywords = journalProfileMap[jourTitle]
				newDoc.JName = jourTitle
			}
		}

		switch area {
		case "EU":
			cat.EU = append(cat.EU, newDoc)
		case "US":
			cat.US = append(cat.US, newDoc)
		case "CN":
			cat.CN = append(cat.CN, newDoc)
		}

	}
	_, err = dataArray.Token()
	if err != nil {
		log.Fatal(err)
	}

}

func buildProfileMap() {

	openFile, _ := os.ReadFile("journalProfiles.json")

	var f interface{}
	err := json.Unmarshal(openFile, &f)

	if err != nil {
		log.Fatal(err)
	}

	m := f.(map[string]interface{})

	for _, v := range m {
		switch vv := v.(type) {
		case []interface{}:
			for _, u := range vv {
				switch vvv := u.(type) {
				case map[string]interface{}:
					for _, g := range vvv {
						switch gg := g.(type) {
						case map[string]interface{}:
							for _, cxx := range gg {
								switch ggx := cxx.(type) {
								case []interface{}:
									for _, gff := range ggx {
										switch fff := gff.(type) {
										case map[string]interface{}:
											title, ok := fff["dc:title"].(string)
											if ok {
												journalProfileMap[title] = fff["subject-area"]
											}
										}
									}
								}
							}

						}
					}
				}
			}
		}
	}
}

func buildQuoteMap() {

	openFile, _ := os.ReadFile("citedby.json")

	var f interface{}
	err := json.Unmarshal(openFile, &f)

	if err != nil {
		log.Fatal(err)
	}

	m := f.(map[string]interface{})

	for _, v := range m {
		switch vv := v.(type) {
		case []interface{}:
			for _, u := range vv {

				switch vtype := u.(type) {
				case map[string]interface{}:
					for k, w := range vtype {
						doi := k
						quoteCount, ok := w.(float64)

						if ok {

						}

						quoteMap[doi] = quoteCount

					}

				}
			}

		}
	}

}

func printResult() {

	var export pandoDBfile

	export.Name = "Citations_Keywoard_Area_Weighted"
	export.Date = time.Now().String()
	export.Id = export.Name + export.Date
	export.Content = cat

	data, err := json.Marshal(export)

	if err != nil {
		log.Fatal(err)
	}

	err2 := os.WriteFile("Citations_Keywoard_Area_Weighted-2.json", data, 0777)

	if err2 != nil {
		log.Fatal(err)
	}

}

func main() {

	buildProfileMap()
	buildQuoteMap()
	buildFile("EU")
	buildFile("CN")
	buildFile("US")
	printResult()

}
