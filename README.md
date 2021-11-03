# stalefish

[![build](https://github.com/kotaroooo0/stalefish/actions/workflows/build.yml/badge.svg)](https://github.com/kotaroooo0/stalefish/actions/workflows/build.yml)
[![test](https://github.com/kotaroooo0/stalefish/actions/workflows/test.yml/badge.svg)](https://github.com/kotaroooo0/stalefish/actions/workflows/test.yml)

stalefish is a toy full text search engine written in Go.
MySQL is used for data persistence now.
Document has only one field.

## Specification

- Indexing Documents
- Search by MatchAllQuery
- Search by MatchQuery(AND,OR)
- Search by PhraseQuery
- Multiple types of analyzers

## Setup

```sh
# Setup MySQL
$ docker-compose up

# Test
$ make test
```

## Example

```go
package main

import (
	"fmt"

	"github.com/kotaroooo0/stalefish"
)

// NOTE: Setup MySQL before execution!
func main() {
	// create index
	db, _ := stalefish.NewDBClient(stalefish.NewDBConfig("root", "password", "127.0.0.1", "3306", "stalefish")) // omit error handling
	storage := stalefish.NewStorageRdbImpl(db)
	analyzer := stalefish.NewAnalyzer([]stalefish.CharFilter{}, stalefish.NewStandardTokenizer(), []stalefish.TokenFilter{stalefish.NewLowercaseFilter()})

	indexer := stalefish.NewIndexer(storage, analyzer, 1)
	for _, body := range []string{"Ruby PHP JS", "Go Ruby", "Ruby Go PHP", "Go PHP"} {
		indexer.AddDocument(stalefish.NewDocument(body))
	}

	// search documents
	sorter := stalefish.NewTfIdfSorter(storage)
	mq := stalefish.NewMatchQuery("GO Ruby", stalefish.OR, analyzer, sorter)
	mseacher := mq.Searcher(storage)
	result, _ := mseacher.Search() // omit error handling
	fmt.Println(result) // [{2 Go Ruby 2} {3 Ruby Go PHP 3} {4 Go PHP 2} {1 Ruby PHP JS 3}]

	// search documents
	pq := stalefish.NewPhraseQuery("go RUBY", analyzer, nil)
	pseacher := pq.Searcher(storage)
	result, _ = pseacher.Search() // omit error handling
	fmt.Println(result) // [{2 Go Ruby 2}
}

```

## Development Task

- [x] Scoring with TF/IDF
- [x] Sorting
- [ ] Setting document fields
- [ ] Replacing MySQL with another DB
- [ ] Preformance Tuning

## Author

kotaroooo0

## LICENSE

MIT
