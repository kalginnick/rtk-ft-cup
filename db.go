package main

import (
	"log"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
	"github.com/blevesearch/bleve/search/query"
)

type DB struct {
	index     bleve.Index
	batch     *bleve.Batch
	batchStep int
	data      map[string]interface{}
}

func NewDB(mapping mapping.IndexMapping) (*DB, error) {
	idx, err := bleve.NewMemOnly(mapping)
	if err != nil {
		return nil, err
	}

	return &DB{index: idx, batch: idx.NewBatch(), data: make(map[string]interface{})}, nil
}

func (db *DB) Index(id string, data interface{}) error {
	db.data[id] = data
	db.batchStep++
	err := db.batch.Index(id, data)
	if err != nil {
		return err
	}
	if db.batchStep > 1000 {
		return db.FlushIndex()
	}
	return nil
}

func (db *DB) FlushIndex() error {
	start := time.Now()
	err := db.index.Batch(db.batch)
	if err != nil {
		return err
	}
	log.Printf("Batch finished in %f seconds\n", time.Now().Sub(start).Seconds())
	db.batch = db.index.NewBatch()
	db.batchStep = 0
	return nil
}

func (db *DB) Search(q query.Query, limit, offset int, asc bool, order ...string) (Response, error) {
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

type Response struct {
	TotalItems uint64 `json:"total_items"`
	Items      []Item `json:"items"`
}

type Item struct {
	Type  string `json:"type"`
	Media *Media `json:"media_item,omitempty"`
	Epg   *Epg   `json:"epg,omitempty"`
}
