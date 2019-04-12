package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

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
		startTime, _ := strconv.Atoi(params.Get("start_time"))
		endTime, _ := strconv.Atoi(params.Get("end_time"))
		//channel, _ := url.QueryUnescape(params.Get("channel_ids"))
		//channels := strings.Split(channel, ",")

		var q strings.Builder
		if startTime > 0 {
			q.WriteString(fmt.Sprintf(" start_time:>=%d ", startTime))
		}
		if endTime > 0 {
			q.WriteString(fmt.Sprintf(" end_time:<=%d ", endTime))
		}
		//for _, _ := range channels {
		//q.WriteString(fmt.Sprintf("channel.id:%s", ch))
		//}
		//q.WriteString(strings.Join(channels, " "))

		result, err := db.Search(q.String(), limit, offset, true)
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

	mediaMapping := bleve.NewDocumentStaticMapping()
	mediaMapping.AddFieldMappingsAt("name", textFieldMapping)
	mediaMapping.AddFieldMappingsAt("description", textFieldMapping)
	mediaMapping.AddFieldMappingsAt("year", numFieldMapping)
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
