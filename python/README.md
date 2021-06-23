# Accessing the Cantabular extended API with Python

The API is powered by [GraphQL](https://graphql.org/) which allows users to request only the data that's relevant to them.

The user is presented with a [GraphiQL](https://github.com/graphql/graphiql) session when the API is accessed via a web browser. GraphiQL is the GraphQL integrated development environment (IDE).

You can write queries in the GraphQL query language and submit these to the API URL. The responses will be in JSON format.

This repository provides some examples on accessing the API with Python.

## Scripts

The scripts demonstrate basic tasks that can be performed using the extended API, but do not exercise the full range of available functionality. They can be used as a starting point for users who want to develop more elaborate applications.

They should be used with `Python 3` and have a dependency on the `requests` library. This should be installed using `pip` if not already present e.g.:

```
pip3 install requests
```

You can see help text by running any of the scripts with a `-h` flag e.g.:

```
python3 variables.py -h
```

### `variables.py`

`variables.py` prints out all the variables in a dataset available for queries. The program separates the output into **rule variables** and the remainder. A rule variable is one which Cantabular uses for per category publication according to disclosure control rules.

This example lists the variables for a locally hosted dataset called `Example`:

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

This example performs a cross tabulation for a locally hosted dataset called `Example` using the variables `country` and `siblings_3`:

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
