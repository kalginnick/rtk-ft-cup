package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/lang/ru"
	"github.com/blevesearch/bleve/mapping"
)

const EpgTypeName = "Epg"

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

func (Epg) Type() string {
	return EpgTypeName
}

type Chanel struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Logo string `json:"logo"`
}

func EpgHandler(db *DB) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		params := request.URL.Query()
		limit, _ := strconv.Atoi(params.Get("limit"))
		if limit == 0 {
			limit = 10
		}
		offset, _ := strconv.Atoi(params.Get("offset"))
		yearGE, _ := strconv.Atoi(params.Get("year_ge"))
		yearLE, _ := strconv.Atoi(params.Get("year_le"))
		//genres, _ := url.QueryUnescape(params.Get("genres"))
		//country, _ := url.QueryUnescape(params.Get("countries"))
		//countries := strings.Split(country, ",")
		sorts, _ := url.QueryUnescape(params.Get("sort_by"))
		sortBy := strings.Split(sorts, ",")
		sortDir := params.Get("sort_dir")

		var q strings.Builder
		if yearGE > 0 {
			q.WriteString(fmt.Sprintf(" year:<=%d ", yearGE))
		}
		if yearLE > 0 {
			q.WriteString(fmt.Sprintf(" year:>=%d ", yearLE))
		}

		result, err := db.Search(q.String(), limit, offset, sortDir != "desc", sortBy...)
		if err != nil {
			ResponseError(writer, 500, err)
		}

		ResponseResult(writer, result)
	})
}

func BuildEpgMapping() *mapping.DocumentMapping {
	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Analyzer = ru.AnalyzerName

	epgMapping := bleve.NewDocumentStaticMapping()
	epgMapping.AddFieldMappingsAt("name", textFieldMapping)
	epgMapping.AddFieldMappingsAt("genre", textFieldMapping)
	epgMapping.AddFieldMappingsAt("description", textFieldMapping)
	channelMapping := bleve.NewDocumentStaticMapping()
	channelMapping.AddFieldMappingsAt("name", textFieldMapping)
	epgMapping.AddSubDocumentMapping("channel", channelMapping)
	return epgMapping
}

func BuildEpgIndex(db *DB, path string) error {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	var objects []Epg
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
