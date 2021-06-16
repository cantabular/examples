// This example uses a live Cantabular instance and dataset which you can
// browse at https://ireland-census-preview.cantabular.com
//
// The API can also be explored using an interactive IDE at
// https://api.ireland-census-preview.cantabular.com/graphql

//
// Remove this line if using in the browser
//
const fetch = require('node-fetch');

const ENDPOINT_URL = 'https://api.ireland-census-preview.cantabular.com/graphql';
const DATASET = 'Ireland-1911-preview';

// Set up a template GraphQL query using parameterised GraphQL variables
// for the dataset, query variables and filters
const QUERY = `
  query($dataset: String!, $variables: [String!]!, $filters: [Filter!]) {
    dataset(name: $dataset) {
      table(variables: $variables, filters: $filters) {
        dimensions {
          count
          categories {
            code
            label
          }
          variable {
            name
            label
          }
        }
        values
      }
    }
  }
`.replace(/\s+/g, ' ');

async function queryCantabularGraphQL(url, query, variables) {
  return fetch(url, {
    method: 'POST',
    mode: 'cors',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      'query': query,
      'variables': variables,
    }),
  })
  .then((response) => {
    if (!response.ok) throw new Error(`${response.status} ${response.statusText}`);
    if (response.status === 204 || response.status === 205) return;
    return response.json();
  });
}

function processCounts(table) {
  const dimLengths = table.dimensions.map((d) => d.count);
  const dimIndices = table.dimensions.map(() => 0);
  let result = [];
  for (let i = 0; i < table.values.length; i++) {
    result.push(populateRow(table, dimIndices, i));
    let j = dimIndices.length - 1;
    while (j >= 0) {
      dimIndices[j] += 1;
      if (dimIndices[j] < dimLengths[j]) break;
      dimIndices[j] = 0;
      j -= 1;
    }
  }
  return result;
}

function populateRow(table, indices, n) {
  const obj = {};
  indices.forEach((index, i) => {
    const dim = table.dimensions[i];
    obj[dim.variable.label] = dim.categories[index].label;
  });
  obj['count'] = table.values[n];
  return obj;
}

function main() {
  queryCantabularGraphQL(ENDPOINT_URL, QUERY, {
    dataset: DATASET,
    variables: ['province', 'age_78cats', 'sex']
  }).then((result) => {
    if (!result) return;
    // Throw an error if a GraphQL error
    // field is present
    if (result.errors) {
      throw new Error(result.errors[0].message);
    }
    const data = processCounts(result.data.dataset.table);
    // Do stuff with data
    console.log(data);
  });
}

main()
