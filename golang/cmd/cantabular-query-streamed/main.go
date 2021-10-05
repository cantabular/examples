// Copyright 2021 The Sensible Code Company Ltd.
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
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cantabular/examples/cmd/cantabular-query-streamed/jsonstream"
	"github.com/cantabular/examples/cmd/cantabular-query-streamed/table"
)

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
// may be processed as it is received without holding the whole response in memory.
// This is known as "streaming". See usage above or run program for help.
func main() {
	if flag.Parse(); len(flag.Args()) < 2 {
		flag.Usage()
		os.Exit(1)
	}
	defer func() {
		if err := recover(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			os.Exit(1)
		}
	}()
	responseBody := makeRequest(flag.Arg(0), flag.Args()[1:])
	defer func() { _ = responseBody.Close() }()
	graphqlJSONToCSV(responseBody, os.Stdout)
}

// makeRequest constructs the GraphQL query and obtains the response. It panics on error.
func makeRequest(dataset string, vars []string) io.ReadCloser {
	const graphQLQuery = `
query($dataset: String!, $variables: [String!]!, $filters: [Filter!]) {
 dataset(name: $dataset) {
  table(variables: $variables, filters: $filters) {
   dimensions {
    count
    variable { name label }
    categories { code label } }
   values
   error
  }
 }
}`
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	if err := enc.Encode(map[string]interface{}{
		"query": graphQLQuery,
		"variables": map[string]interface{}{
			"dataset":   dataset,
			"variables": vars,
		},
	}); err != nil {
		panic(fmt.Sprintf("Error encoding JSON request body: %s", err))
	}

	resp, err := http.Post(*apiUrl, "application/json", &b)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != http.StatusOK {
		panic(resp.Status)
	}
	return resp.Body
}

// graphqlJSONToCSV converts a JSON response in r to CSV on w and panics on error
func graphqlJSONToCSV(r io.Reader, w io.Writer) {
	dec := jsonstream.New(r)
	if !dec.StartObjectComposite() {
		panic("No JSON object found in response")
	}
	for dec.More() {
		switch field := dec.DecodeName(); field {
		case "data":
			if dec.StartObjectComposite() {
				decodeDataFields(dec, w)
				dec.EndComposite()
			}
		case "errors":
			decodeErrorsPanicIfAny(dec)
		}
	}
	dec.EndComposite()
}

// decodeDataFields decodes the fields of the data part of the GraphQL response, writing CSV to w
func decodeDataFields(dec jsonstream.Decoder, w io.Writer) {
	mustMatchName := func(name string) {
		if gotName := dec.DecodeName(); gotName != name {
			panic(fmt.Sprintf("Expected %q but got %q", name, gotName))
		}
	}
	mustMatchName("dataset")
	if !dec.StartObjectComposite() {
		panic(`dataset object expected but "null" found`)
	}
	mustMatchName("table")
	if dec.StartObjectComposite() {
		decodeTableFields(dec, w)
		dec.EndComposite()
	}
	dec.EndComposite()
}

// decodeErrorsPanicIfAny decodes the errors part of the GraphQL response and
// panics with the error message(s) if there are any.
func decodeErrorsPanicIfAny(dec jsonstream.Decoder) {
	var graphqlErrs []struct{ Message string }
	if err := dec.Decode(&graphqlErrs); err != nil {
		panic(err)
	}
	var sb strings.Builder
	for _, err := range graphqlErrs {
		if sb.Len() > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(err.Message)
	}
	if sb.Len() > 0 {
		panic(sb.String())
	}
}

// decodeTableFields decodes the fields of the table part of the GraphQL response, writing CSV to w.
// If no table cell values are present then no output is written.
func decodeTableFields(dec jsonstream.Decoder, w io.Writer) {
	var dims table.Dimensions
	for dec.More() {
		switch field := dec.DecodeName(); field {
		case "dimensions":
			if err := dec.Decode(&dims); err != nil {
				panic(err)
			}
		case "error":
			if errMsg := dec.DecodeString(); errMsg != nil {
				panic(fmt.Sprintf("Table blocked: %s", *errMsg))
			}
		case "values":
			if dims == nil {
				panic("values received before dimensions")
			}
			if dec.StartArrayComposite() {
				decodeValues(dec, dims, w)
				dec.EndComposite()
			}
		}
	}
}

// decodeValues decodes the values of the cells in the table, writing CSV to w.
func decodeValues(dec jsonstream.Decoder, dims table.Dimensions, w io.Writer) {
	cw := csv.NewWriter(w)
	// csv.Writer errors are sticky, so we only need to check when flushing at the end
	defer func() {
		cw.Flush()
		if err := cw.Error(); err != nil {
			panic(err)
		}
	}()
	// construct the CSV header and write it
	columns := make([]string, 0, len(dims)+1)
	for _, d := range dims {
		columns = append(columns, d.Variable.Label)
	}
	_ = cw.Write(append(columns, "count"))
	// write the data rows
	for ti := dims.NewIterator(); dec.More(); {
		columns = columns[:0] // save allocations
		for i := range dims {
			columns = append(columns, ti.CategoryAtColumn(i).Label)
		}
		_ = cw.Write(append(columns, dec.DecodeNumber().String()))
		ti.Next()
	}
}
