package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/lang/ru"
)

func main() {
	indexMapping := bleve.NewIndexMapping()
	indexMapping.DefaultAnalyzer = ru.AnalyzerName
	indexMapping.AddDocumentMapping("Media", BuildMediaMapping())
	indexMapping.AddDocumentMapping("Epg", BuildEpgMapping())

	db, err := NewDB(indexMapping)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		err = BuildMediaIndex(db, "testdata/media_test.json")
		if err != nil {
			log.Fatal(err)
		}
	}()

	http.Handle("/api/v1/search", GET(SearchHandler(db)))
	http.Handle("/api/v1/media_items", GET(MediaHandler(db)))
	http.Handle("/api/v1/epg", GET(EpgHandler(db)))
	fmt.Println("Running on http://localhost:8080...")
	http.ListenAndServe(":8080", nil)
}
