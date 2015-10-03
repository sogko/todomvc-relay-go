# todomvc-relay-go
Port of the [React/Relay TodoMVC app](https://github.com/facebook/relay/tree/master/examples/todo), driven by a Golang GraphQL backend

## Parts and pieces
- [golang-relay-starter-kit](https://github.com/sogko/golang-relay-starter-kit)
- [graphql-go](https://github.com/chris-ramon/graphql-go)
- [graphql-go-handler](https://github.com/sogko/graphql-go-handler)
- [graphql-relay-go](https://github.com/sogko/graphql-relay-go)

### Notes:
This is based on alpha version of `graphql-go` and `graphql-relay-go`. 
Be sure to watch both repositories for latest changes.

## Installation

1. Install dependencies for NodeJS app server
```
npm install
```
2. Install dependencies for Golang GraphQL server
```
go get -v ./...
```

## Running

Start a local server:

```
npm start
```

The above command will run both the NodeJS app server and Golang GraphQL server concurrently.

- Golang GraphQL server will be running at http://localhost:8080/graphql
- NodeJS app server will be running at http://localhost:3000

## Developing

Any changes you make to files in the `js/` directory will cause the server to
automatically rebuild the app and refresh your browser.

If at any time you make changes to `data/schema.go`, stop the server,
regenerate `data/schema.json`, and restart the server:

```
npm run update-schema
npm start
```
