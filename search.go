package main

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
)

func SearchHandler(db *DB) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		params := request.URL.Query()
		limit, _ := strconv.Atoi(params.Get("limit"))
		if limit == 0 {
			limit = 10
		}
		offset, _ := strconv.Atoi(params.Get("offset"))
		input, _ := url.QueryUnescape(params.Get("query"))

		result, err := db.Search(toQuery(input), limit, offset, true)
		if err != nil {
			ResponseError(writer, 500, err)
		}

		ResponseResult(writer, result)
	})
}

func toQuery(input string) query.Query {
	terms := notEmpty(strings.Split(strings.ToLower(input), " "))

	queries := make([]query.Query, 1)
	queries[0] = bleve.NewMatchPhraseQuery(strings.Join(terms, " "))
	for _, t := range terms {
		queries = append(queries,
			bleve.NewFuzzyQuery(t),
			bleve.NewMatchQuery(t),
			bleve.NewMatchQuery(fixEnRu(t)),
			bleve.NewMatchQuery(fixTranslit(t)),
		)
	}
	return bleve.NewDisjunctionQuery(queries...)
}

func notEmpty(s []string) []string {
	filtered := s[:0]
	for _, t := range s {
		if tmp := strings.TrimSpace(t); tmp != "" {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

func fixEnRu(s string) string {
	var buf strings.Builder
	for _, r := range s {
		if i, ok := enru[r]; ok {
			buf.WriteRune(i)
		}
	}
	return buf.String()
}

func fixTranslit(s string) string {
	s = translitReplacer.Replace(s)
	for _, r := range s {
		if i, ok := translitRu[r]; ok {
			s = strings.ReplaceAll(s, string(r), i)
		}
	}
	return s
}

var translitReplacer = strings.NewReplacer("ch", "ч", "sh", "ш", "sch", "щ", "kh", "х", "zh", "ж", "yu", "ю")

var translitRu = map[rune]string{
	'f': "ф",
	'i': "и",
	's': "с",
	'v': "в",
	'u': "у",
	'a': "а",
	'h': "х",
	'p': "п",
	'r': "р",
	'o': "о",
	'l': "л",
	'd': "д",
	't': "т",
	'z': "з",
	'k': "к",
	'e': "е",
	'g': "г",
	'm': "м",
	'c': "ц",
	'n': "н",
	'y': "я",
	'b': "б",
}

var enru = map[rune]rune{
	'a': 'ф',
	'b': 'и',
	'c': 'с',
	'd': 'в',
	'e': 'у',
	'f': 'а',
	'g': 'п',
	'h': 'р',
	'i': 'ш',
	'j': 'о',
	'k': 'л',
	'l': 'д',
	'm': 'ь',
	'n': 'т',
	'o': 'щ',
	'p': 'з',
	'r': 'к',
	's': 'ы',
	't': 'е',
	'u': 'г',
	'v': 'м',
	'w': 'ц',
	'x': 'ч',
	'y': 'н',
	'z': 'я',
	'[': 'х',
	']': 'ъ',
	';': 'ж',
	',': 'б',
	'.': 'ю',
}
