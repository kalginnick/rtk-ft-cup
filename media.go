package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/blevesearch/bleve/analysis/analyzer/keyword"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/lang/ru"
	"github.com/blevesearch/bleve/mapping"
)

const MediaTypeName = "Media"

type Media struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Kind        string   `json:"type"`
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

func (Media) Type() string {
	return MediaTypeName
}

type Person struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func MediaHandler(db *DB) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		params := request.URL.Query()
		limit, _ := strconv.Atoi(params.Get("limit"))
		if limit == 0 {
			limit = 10
		}
		offset, _ := strconv.Atoi(params.Get("offset"))
		yearFrom, _ := strconv.ParseFloat(params.Get("year_ge"), 64)
		yearTo, _ := strconv.ParseFloat(params.Get("year_le"), 64)
		//genres, _ := url.QueryUnescape(params.Get("genres"))
		//country, _ := url.QueryUnescape(params.Get("countries"))
		//countries := strings.Split(country, ",")
		sorts, _ := url.QueryUnescape(params.Get("sort_by"))
		sortBy := strings.Split(sorts, ",")
		sortDir := params.Get("sort_dir")

		year := bleve.NewNumericRangeQuery(&yearFrom, &yearTo)
		year.FieldVal = "year"
		result, err := db.Search(bleve.NewDisjunctionQuery(year), limit, offset, sortDir != "desc", sortBy...)
		if err != nil {
			ResponseError(writer, 500, err)
		}

		ResponseResult(writer, result)
	})
}

func BuildMediaMapping() *mapping.DocumentMapping {
	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Analyzer = ru.AnalyzerName
	numFieldMapping := bleve.NewNumericFieldMapping()
	keywordFieldMapping := bleve.NewTextFieldMapping()
	keywordFieldMapping.Analyzer = keyword.Name

	mediaMapping := bleve.NewDocumentStaticMapping()
	mediaMapping.AddFieldMappingsAt("name", textFieldMapping)
	mediaMapping.AddFieldMappingsAt("description", textFieldMapping)
	mediaMapping.AddFieldMappingsAt("year", numFieldMapping)
	mediaMapping.AddFieldMappingsAt("start_time", numFieldMapping)
	mediaMapping.AddFieldMappingsAt("end_time", numFieldMapping)
	mediaMapping.AddFieldMappingsAt("countries", textFieldMapping)
	mediaMapping.AddFieldMappingsAt("genres", textFieldMapping)
	personMapping := bleve.NewDocumentStaticMapping()
	personMapping.AddFieldMappingsAt("name", textFieldMapping)
	mediaMapping.AddSubDocumentMapping("persons", personMapping)
	return mediaMapping
}

func BuildMediaIndex(db *DB, path string) error {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	var objects []Media
	err = json.Unmarshal(buf, &objects)
	if err != nil {
		return err
	}
	for _, o := range objects {
		err = db.Index(strconv.FormatInt(o.ID, 10), o)
		if err != nil {
			return err
		}
	}

	return db.FlushIndex()
}
