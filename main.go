package main

import (
	"log"
	"net/http"
	"os"
	"path"

	"github.com/gorilla/handlers"

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
	dataDir := os.Getenv("DATA_DIR")
	go func() {
		err = BuildMediaIndex(db, path.Join(dataDir, "media_items.json"))
		if err != nil {
			log.Fatal(err)
		}
	}()
	//go func() {
	//err = BuildEpgIndex(db, path.Join(dataDir, "epg.json"))
	//if err != nil {
	//	log.Fatal(err)
	//}
	//}()

	http.Handle("/api/v1/search", handlers.CORS()(GET(SearchHandler(db))))
	http.Handle("/api/v1/media_items", GET(MediaHandler(db)))
	http.Handle("/api/v1/epg", GET(EpgHandler(db)))
	log.Print("Running on http://localhost:8080...")
	http.ListenAndServe(":8080", nil)
}
