# basic-graphql-federation

Basic GraphQL federation 2 example using Apollo Router (Rust-based binary) and Go services written with [gqlgen](https://gqlgen.com/).

Run `docker compose up` to run entire stack. For local development see each service individual README.

Open `http://localhost:4000` to open Apollo Explorer and query:

```graphql
query topProducts($limit: Int) {
  topProducts(limit: $limit) {
    title
    description
    products {
      sku
      name
      price
      description
    }
  }
}
```
