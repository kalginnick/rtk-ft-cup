package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/lang/ru"
	"github.com/blevesearch/bleve/mapping"
)

func main() {
	indexMapping := bleve.NewIndexMapping()
	mediaMapping := bleve.NewDocumentMapping()
	indexMapping.AddDocumentMapping("Media", mediaMapping)

	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Analyzer = ru.AnalyzerName

	numFieldMapping := bleve.NewNumericFieldMapping()

	mediaMapping.AddFieldMappingsAt("name", textFieldMapping)
	mediaMapping.AddFieldMappingsAt("description", textFieldMapping)
	mediaMapping.AddFieldMappingsAt("year", numFieldMapping)

	db, err := NewDB(indexMapping)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		err = db.BuildIndex("testdata/media_test.json")
		if err != nil {
			log.Fatal(err)
		}
	}()

	http.Handle("/api/v1/search", GET(SearchHandler(db)))
	http.Handle("/api/v1/media_items", GET(MediaItemsHandler(db)))
	http.Handle("/api/v1/epg", GET(EpgHandler(db)))
	fmt.Println("Running on :8080...")
	http.ListenAndServe(":8080", nil)
}

func GET(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			ResponseError(res, http.StatusMethodNotAllowed, nil)
			return
		}
		next.ServeHTTP(res, req)
	})
}

func ResponseError(w http.ResponseWriter, statusCode int, err error) {
	if err != nil {
		log.Print(err)
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(statusCode)
	io.WriteString(w, http.StatusText(statusCode))
}

type DB struct {
	index bleve.Index
	data  map[string]interface{}
}

func NewDB(mapping mapping.IndexMapping) (*DB, error) {
	idx, err := bleve.NewMemOnly(mapping)
	if err != nil {
		return nil, err
	}

	return &DB{index: idx, data: make(map[string]interface{})}, nil
}

func (db *DB) Search(query string, limit, offset int, asc bool, order ...string) (Response, error) {
	q := bleve.NewQueryStringQuery(fmt.Sprintf("%s", query))
	request := bleve.NewSearchRequestOptions(q, limit, offset, true)
	if len(order) > 0 {
		if !asc {
			for i, o := range order {
				order[i] = "-" + o
			}
		}
		request.SortBy(order)
	}
	result, err := db.index.Search(request)
	if err != nil {
		return Response{}, err
	}

	response := Response{
		TotalItems: result.Total,
		Items:      make([]Item, len(result.Hits)),
	}
	for i, hit := range result.Hits {
		object := db.data[hit.ID]
		switch item := object.(type) {
		case Media:
			response.Items[i].Type = "media_item"
			response.Items[i].Media = &item
		case Epg:
			response.Items[i].Type = "epg"
			response.Items[i].Epg = &item
		}
	}

	return response, nil
}

func (db *DB) BuildIndex(sourcePath string) error {
	r, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer r.Close()

	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	var medias []Media
	err = json.Unmarshal(buf, &medias)
	if err != nil {
		return err
	}
	last := len(medias) - 1
	fmt.Printf("%d JSON items decoded\n", last+1)
	batch := db.index.NewBatch()
	var n int
	for i, m := range medias {
		id := strconv.FormatInt(m.ID, 10)
		db.data[id] = m
		err = batch.Index(id, m)
		if err != nil {
			return err
		}
		if i > 1000 || i == last {
			n++
			start := time.Now()
			err = db.index.Batch(batch)
			if err != nil {
				return err
			}
			fmt.Printf("Batch %d finished in %f seconds\n", n, time.Now().Sub(start).Seconds())
			batch = db.index.NewBatch()
			i = 0
		}
	}

	return nil
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

func MediaItemsHandler(db *DB) http.Handler {
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

func ResponseResult(writer http.ResponseWriter, response Response) {
	buf, err := json.Marshal(response)
	if err != nil {
		ResponseError(writer, 500, err)
	}

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprint(writer, string(buf))
}

func SearchHandler(db *DB) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		params := request.URL.Query()
		limit, _ := strconv.Atoi(params.Get("limit"))
		if limit == 0 {
			limit = 10
		}
		offset, _ := strconv.Atoi(params.Get("offset"))
		query, _ := url.QueryUnescape(params.Get("query"))

		result, err := db.Search(query, limit, offset, true)
		if err != nil {
			ResponseError(writer, 500, err)
		}

		ResponseResult(writer, result)
	})
}

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

type Response struct {
	TotalItems uint64 `json:"total_items"`
	Items      []Item `json:"items"`
}

type Item struct {
	Type  string `json:"type"`
	Media *Media `json:"media_item,omitempty"`
	Epg   *Epg   `json:"epg,omitempty"`
}
