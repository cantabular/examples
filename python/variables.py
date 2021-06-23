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

from argparse import ArgumentParser
import json
import requests

GRAPHQL_QUERY = """
query($dataset: String!) {
 dataset(name: $dataset) {
  variables {
  edges {
   node {
    name
   }
  }
 }
 ruleBase {
  name
  isSourceOf {
   edges {
     node {
      name
     }
    }
   }
  }
 }
}
"""


def main():
    """
    Perform a GraphQL query to identify dataset variables.

    Query the Cantabular extended API, verify and parse the response
    and print the names of variables in the dataset.
    """
    args = parse_arguments()

    graphql_variables = json.dumps({
        "dataset": args.dataset,
    })
    http_resp = requests.post(args.base_url, data={"query": GRAPHQL_QUERY,
                                                   "variables": graphql_variables})
    http_resp.raise_for_status()

    resp = http_resp.json()
    if "errors" in resp:
        raise RuntimeError(resp["errors"][0]["message"])

    # In Cantabular "rule variables" are special variables used for rule based redaction of
    # categories when a cross tabulation is generated. Every rule variable is derived from
    # the same base variable. Not all datasets have rule based redaction of categories. However,
    # rule variables are still handled differently (e.g. by the UI) and a query may contain at
    # most one rule variable which must be the first variable specified in a cross tabulation
    # query.
    rule_variables = {v["node"]["name"] for v in
                      resp["data"]["dataset"]["ruleBase"]["isSourceOf"]["edges"]}
    print("RULE VARIABLES")
    print("--------------")
    for variable in rule_variables:
        print(variable)

    variables = {v["node"]["name"] for v in resp["data"]["dataset"]["variables"]["edges"]}
    print("")
    print("OTHER VARIABLES")
    print("---------------")
    non_rule_variables = variables - rule_variables
    for variable in non_rule_variables:
        print(variable)


def parse_arguments():
    """Parse command line arguments."""
    parser = ArgumentParser(description="Identify the variables in a dataset via the Cantabular "
                            "extended API using the GraphQL endpoint. Verify and parse the "
                            "response and print the variable names to screen. Variables are "
                            "separated into rule variables (which are used for rule based "
                            "redaction of categories in cross tabulations), and other variables.")

    # Positional arguments
    parser.add_argument("base_url",
                        metavar="URL",
                        type=str,
                        help="Cantabular extended API base URL "
                             "e.g. http://localhost:8492/graphql")

    parser.add_argument("dataset",
                        metavar="Dataset",
                        type=str,
                        help="Name of dataset to use in query")

    return parser.parse_args()


if __name__ == "__main__":
    main()
