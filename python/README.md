# Accessing the Cantabular extended API with Python

The API is powered by [GraphQL](https://graphql.org/) which allows users to request only the data that's relevant to them.

The user is presented with a [GraphiQL](https://github.com/graphql/graphiql) session when the API is accessed via a web browser. GraphiQL is the GraphQL integrated development environment (IDE).

Requests to the API via other means return reponses in GraphQL format, which is a little like JSON.

This repository provides some examples on accessing the API with Python.

## Scripts

The scripts demonstrate basic tasks that can be performed using the extended API, but do not exercise the full range of available functionality. They can be used as a starting point for users who want to develop more elaborate applications.

They should be used with `Python 3` and have a dependency on the `requests` library. This should be installed using `pip` if not already present e.g.:

```
pip3 install requests
```

### `variables.py`

`variables.py` is used to identify the variables associated with a specific dataset. It constructs a GraphQL query using a user-supplied base URL and dataset name and prints the names of all the available variables. The variables are separated into **rule variables** and other variables. Rule variables are special variables that are used in rule based redaction of categories when cross tabulations are generated.

The script takes 2 parameters. The first is the base URL for the GraphQL endpoint, whilst the second is the name of the dataset. This example lists the variables for a locally hosted dataset called `Example`:

```
> python3 variables.py http://localhost:8492/graphql Example
RULE VARIABLES
--------------
city
country

OTHER VARIABLES
---------------
siblings_3
siblings
sex
```

### `query.py`

`query.py` is used to perform cross tabulations and print the resulting table in CSV format. It constructs a GraphQL query using a user-supplied base URL, dataset name and variable list.


The script takes at least 3 parameters. The first is the base URL for the GraphQL endpoint, whilst the second is the name of the dataset. All subsequent parameters are treated as the names of query variables. At most one rule variable can be specified and this must be the first supplied variable. This example performs a cross tabulation for a locally hosted dataset called `Example` using the variables `country` and `siblings_3`:

```
> python3 query.py http://localhost:8492/graphql Example country siblings_3
Country,Number of siblings (3 mappings),count
England,No siblings,1
England,1 or 2 siblings,0
England,3 or more siblings,2
Northern Ireland,No siblings,0
Northern Ireland,1 or 2 siblings,1
Northern Ireland,3 or more siblings,2
```
