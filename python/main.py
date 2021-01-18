"""
Copyright 2020 The Sensible Code Company Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.

You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
"""

import os
import sys
import itertools
import csv
import requests

ENDPOINT_URL = "http://localhost:8492/graphql"

GRAPHQL_QUERY = """
{
 dataset(name: "Example") {
  table(variables: ["city", "siblings"]) {
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
}
"""


def main():
    """
    Perform a GraphQL query and handle response.

    Query the Cantabular extended API, verify and parse the response
    and print the rows containing in the table in CSV format.
    """
    http_resp = requests.post(ENDPOINT_URL, data={"query": GRAPHQL_QUERY})
    http_resp.raise_for_status()

    resp = http_resp.json()
    if "errors" in resp:
        raise RuntimeError(resp["errors"][0]["message"])

    table = resp["data"]["dataset"]["table"]

    if table["error"] is not None:
        raise RuntimeError(table["error"])

    writer = csv.writer(sys.stdout, lineterminator=os.linesep)

    # Write the table header.
    dimension_labels = [d["variable"]["label"] for d in table["dimensions"]]
    writer.writerow(dimension_labels + ["count"])

    dimension_categories = [d["categories"] for d in table["dimensions"]]

    # Create an iterator that will generate all combinations of dimension
    # categories. The combinations are generated in row-major order which
    # matches the order of the table "values".
    iterator = itertools.product(*dimension_categories)

    # Iterate over the category combinations and display the code for each
    # category along with the corresponding value.
    # A maximum number of values to return can be specified in the query so
    # use itertools.islice to restrict the number of category combinations
    # that the iterator returns.
    max_rows = len(table["values"])
    for i, cats in enumerate(itertools.islice(iterator, max_rows)):
        columns = [c["label"] for c in cats]
        columns.append(table["values"][i])
        writer.writerow(columns)


if __name__ == "__main__":
    main()
