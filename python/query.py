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
import json
from argparse import ArgumentParser
import requests

GRAPHQL_QUERY = """
query($dataset: String!, $variables: [String!]!) {
 dataset(name: $dataset) {
  table(variables: $variables) {
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
    and print the rows contained in the table in CSV format.
    """
    args = parse_arguments()

    graphql_variables = json.dumps({
        "dataset": args.dataset,
        "variables": args.variables,
    })
    http_resp = requests.post(args.base_url, data={"query": GRAPHQL_QUERY,
                                                   "variables": graphql_variables})
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


def parse_arguments():
    """Parse command line arguments."""
    parser = ArgumentParser(description="Query the Cantabular extended API using the GraphQL "
                            "endpoint. Verify and parse the response and print the rows contained "
                            "in the table in CSV format.")

    # Positional arguments
    parser.add_argument("base_url",
                        metavar="URL",
                        type=str,
                        help="Cantabular GraphQL base URL including port "
                             "e.g. http://localhost:8492/graphql")

    parser.add_argument("dataset",
                        metavar="Dataset",
                        type=str,
                        help="Name of dataset to use in query")

    parser.add_argument("variables",
                        metavar="Variables",
                        nargs="+",
                        type=str,
                        help="Names of variables to use in query. At least one variable name "
                        "must be supplied. A single rule variable may be specified. If a rule "
                        "variable is specified then it must be the first variable.")

    return parser.parse_args()


if __name__ == "__main__":
    main()
