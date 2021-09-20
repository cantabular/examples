// Copyright 2020 The Sensible Code Company Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
//
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// For function see description of main() method.
package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type (
	Response struct {
		Data struct {
			Dataset struct {
				Table Table
			}
		}

		Errors []struct {
			Message string
		}
	}

	Table struct {
		Dimensions []struct {
			Count      int
			Categories []Category
			Variable   Variable
		}

		Values []int
		Error  string
	}

	Variable struct {
		Name, Label string
	}

	Category struct {
		Code, Label string
	}

	Row struct {
		Categories []Category
		Count      int
	}
)

// ForEachRow calls the provided function for each row of the returned data.
//
// Panics if the table contains an error.
func (t Table) ForEachRow(cb func(row *Row)) {
	if t.Error != "" {
		panic(t.Error)
	}

	numDimensions := len(t.Dimensions)

	// first, get a slice containing the length of each dimension:
	dimCounts := make([]int, 0, numDimensions)
	for _, dim := range t.Dimensions {
		dimCounts = append(dimCounts, dim.Count)
	}

	// next, get a slice of equal length containing zeroes.
	dimIndices := make([]int, numDimensions)

	// finally, iterate through the rows and update the indices.
	row := Row{Categories: make([]Category, numDimensions)}

	for i := range t.Values {
		t.populateRow(&row, dimIndices, i)
		cb(&row)

		j := len(dimIndices) - 1
		for j >= 0 {
			dimIndices[j] += 1
			if dimIndices[j] < dimCounts[j] {
				break
			}
			dimIndices[j] = 0
			j -= 1
		}
	}
}

func (t Table) populateRow(row *Row, indices []int, i int) {
	for j, k := range indices {
		dimCat := &t.Dimensions[j].Categories[k]
		rowCat := &row.Categories[j]
		rowCat.Code, rowCat.Label = dimCat.Code, dimCat.Label
	}
	row.Count = t.Values[i]
}

func (t Table) Header() []string {
	result := make([]string, 0, len(t.Dimensions))
	for _, d := range t.Dimensions {
		result = append(result, d.Variable.Label)
	}
	return append(result, "count")
}

const graphQLQuery = `
query($dataset: String!, $variables: [String!]!, $filters: [Filter!]) {
 dataset(name: $dataset) {
  table(variables: $variables, filters: $filters) {
   dimensions {
    count
    variable {
     name
     label
    }
    categories {
     code
     label
    }
   }
   values
   error
  }
 }
}`

var apiUrl = flag.String("u", "http://localhost:8492/graphql",
	"Extended API URL")

func init() {
	const usage = `Usage: %s <dataset-name> <var> [<var> ...]

Writes table output to stdout as CSV.
Exit code is one on error and errors are reported to stderr.

Options:
`
	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), usage, filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
}

// This example demonstrates how tabulated data returned via a GraphQL request
// may be processed in the simplest way possible using decoding into a Go type.
// See usage above or run program for help.
func main() {
	if flag.Parse(); len(flag.Args()) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	if err := enc.Encode(map[string]interface{}{
		"query": graphQLQuery,
		"variables": map[string]interface{}{
			"dataset":   flag.Arg(0),
			"variables": flag.Args()[1:],
		},
	}); err != nil {
		log.Fatalf("Error encoding JSON request body: %s", err)
	}

	resp, err := http.Post(*apiUrl, "application/json", &b)
	if err != nil {
		log.Fatal(err)
	}

	// Decode the response.
	var gqlResp Response
	if err = json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		log.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check for table errors
	if len(gqlResp.Errors) > 0 {
		log.Fatalf("Unexpected error: %v", gqlResp.Errors)
	}
	table := gqlResp.Data.Dataset.Table

	// Iterate through each row, and print it:
	cw := csv.NewWriter(os.Stdout)
	defer func() {
		cw.Flush()
		if err := cw.Error(); err != nil {
			log.Fatal(err)
		}
	}()
	// csv.Writer errors are sticky: log in defer
	_ = cw.Write(table.Header())

	var columns []string
	table.ForEachRow(func(row *Row) {
		columns = columns[:0]
		for i := range row.Categories {
			columns = append(columns, row.Categories[i].Label)
		}
		_ = cw.Write(append(columns, strconv.Itoa(row.Count)))
	})
}
