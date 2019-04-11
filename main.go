package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
)

func main() {
	indexMapping := bleve.NewIndexMapping()
	mediaMapping := bleve.NewDocumentMapping()
	indexMapping.AddDocumentMapping("media", mediaMapping)
	mediaNameFieldMapping := bleve.NewTextFieldMapping()
	mediaMapping.AddFieldMappingsAt("name", mediaNameFieldMapping)

	//ergs := Index("./testdata/epg.json", mapping)
	medias := Index("testdata/media_items.json", indexMapping)

	//http.Handle("/api/v1/search", IndexHandler(SearchHandler, medias, ergs))
	http.Handle("/api/v1/media_items", IndexHandler(MediaItemsHandler, medias))
	//http.Handle("/api/v1/epg", IndexHandler(EpgHandler, ergs))
	fmt.Println("Runnning...")
	http.ListenAndServe(":8080", nil)
}

func IndexHandler(handler http.HandlerFunc, index ...bleve.Index) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		handler.ServeHTTP(writer, request)
	})
}

func Index(path string, mapng mapping.IndexMapping) bleve.Index {
	index, err := bleve.New(path+".ind", mapng)
	if err != nil {
		panic(err)
	}

	r, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	buf, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	var medias = make([]Media, 0)
	err = json.Unmarshal(buf, &medias)
	if err != nil {
		panic(err)
	}
	fmt.Println("Json decoded")
	batch := index.NewBatch()
	var count int
	for i, m := range medias {
		err = batch.Index(strconv.FormatInt(m.ID, 10), m)
		count++
		if err != nil {
			panic(err)
		}
		if count > 100 {
			err = index.Batch(batch)
			if err != nil {
				panic(err)
			}
			batch = index.NewBatch()
			count = 0
		}
		fmt.Println(i)
	}

	return index
}

func EpgHandler(writer http.ResponseWriter, request *http.Request) {

}

func MediaItemsHandler(writer http.ResponseWriter, request *http.Request) {

}

func SearchHandler(writer http.ResponseWriter, request *http.Request) {

}

type Media struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Duration    int      `json:"duration"`
	Countries   []string `json:"countries"`
	AgeValue    int      `json:"age_value"`
	Year        string   `json:"year"`
	Logo        string   `json:"logo"`
	Rating      float32  `json:"rating"`
	Description string   `json:"description"`
	Genres      []string `json:"genres"`
	Persons     []Person `json:"persons"`
	Packages    []int64  `json:"packages"`
	AssetTypes  []int64  `json:"asset_types"`
}

type Person struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Epg struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	AgeValue    int    `json:"age_value"`
	StartTime   int    `json:"start_time"`
	EndTime     int    `json:"end_time"`
	Genre       string `json:"genre"`
	Description string `json:"description"`
	Logo        string `json:"logo"`
	Channel     Chanel `json:"channel"`
	LocationID  int    `json:"location_id"`
}

type Chanel struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Logo string `json:"logo"`
}

type Response struct {
	TotalItems int `json:"total_itens"`
	Items      []Media
}
