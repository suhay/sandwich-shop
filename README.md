# Sandwich Shop
## A containerless, serverless experiment using Go and GraphQL

*Note! This is purely experimental and not meant for production use.*

The purpose of this experiment is to explore other alternatives to the container-based, serverless pattern. We are hoping this solution will minimize the overhead needed for spinning up a new container on-demand and fully eliminate the idea of a "cold start”.

## Usage

### Build

```bash
gh repo clone suhay/sandwich-shop
cd sandwich-shop
make build
```

This will build directly from the source code. The compiled binary will be within the project directory. One benefit of doing it this way is the `.env` file path will default to the project's root. This means you will not need to specify it as an argument.

### Install

```bash
go install github.com/suhay/sandwich-shop@latest
go install github.com/suhay/sandwich-shop/sandwiches/gowich@latest
```

Install the latest release of Sandwich Shop. This will add the binary to your `GOPATH`. The benefit of installing vs building is being able to run the binary within any directory. You will, however, need to specify a path to the `.env` file you want to use as an argument. Also, be sure to add `$GOPATH/bin` to your `PATH`,

### Run

```bash
sandwich-shop [--env=<path>] [--port=<port>]
```

If you built the application from the source, you will need to execute the binary within the project directory.

| Flag | Description |
| --- | --- |
| --env | Path to a local, or remote .env this process should use for default configurations. |
| --port | The port number to run on. Default: 3002 |

## Setting up Shop

Once you have the Sandwich Shop installed, and you have at least one Sandwich ready to go (`gowich` or `nodewich`), the next task is to set up your `.env` files. These can live anywhere and are included as CLI arguments when launching the Shop or Sandwich.

```bash
# .env

MONGODB_URL=cluster0.mongodb.net
MONGODB_USER=mango
MONGODB_PASSWD=1234password
JWT_SECRET=1234567890jwtsecretcode
TIMEOUT=60
TENANTS=~/sandwich-shop/tenants
```

| Key | Description |
| --- | --- |
| MONGODB_URL | (optional) URL to the mongo database that stores tenant information including the Bearer token. |
| MONGODB_USER | (optional) User for logging into the tenant info database. |
| MONGODB_PASSWD | (optional) Password for the user above. |
| JWT_SECRET | Used for signing a JWT payload while passing it through the shop as an order. This is used to verify the order came through a shop directly and that it was unchanged while in transit. This secret must be identical across all Sandwiches. |
| TIMEOUT | Seconds to wait on the order to complete. |
| TENANTS | The location where your shop tenants are stored. Think of these as project roots, function roots, or user roots of what will be using the Shop. |

### Registering a Sandwich

A Sandwich is what will handle the heavy lifting. Each Sandwich can be specialized in running one type of application (Node.js, Golang, Python, etc.), or they can do a mix of them, whichever you'd prefer. There are two ways you can register your Sandwiches. The first way is through a local `sandiwches.json` file. This is the more clunky way to maintain your Sandwich registry, but it cuts out a round trip to a MongoDB. The second way is through a MongoDB collection.

```json
[
  {
    "_id": "rye",
    "name": "Rye",
    "host": "http://localhost",
    "port": 4006,
    "runtimes": [
      "node12"
    ]
  },
  {
    "_id": "ciabatta",
    "name": "Ciabatta",
    "host": "http://localhost",
    "port": 4007,
    "runtimes": [
      "go1_17",
      "node14",
      "node16",
      "python3"
    ]
  }
]
```

| Key | Description |
| --- | --- |
| _id | Sandwich's id |
| name | Human readable names for easily identifying when there is an error. |
| host | URL the Sandwich will be reached by the Sandwich Shop process (how it will receive orders). |
| port | The port on the designated host the Sandwich is reachable on |
| runtimes | An array of runtime identifiers that show what this Sandwich is set to run. |

Runtimes must be defined within the GraphQL resolver which means adding one will require a pull request. The above schema can also be stored within a MongoDB collection called `sandwiches` on the `sandwich-shop` database you are using to run this application.

### Gowich

A `Gowich` is a Sandwich written in Go. You may have as many of these running as needed as long as they are all running on separate ports. Each `Gowich` should have a `.env` file supplied to it as a CLI argument.

```bash
# .gowich1.env

PORT=3001
TIMEOUT=60
JWT_SECRET=1234567890jwtsecretcode
GO1_13=/usr/local/go/bin/go
TENANTS=/path/to/tenants
```

| Key | Description |
| --- | --- |
| PORT | Port to run this Sandwich on. This will be overridden if --port is supplied as a command argument. |
| TIMEOUT | Timeout, in seconds, to wait for an order to finish. |
| JWT_SECRET | Must be the same JWT_SECRET as specified in the Sandwich Shop’s .env variables. This is used to ensure the request was generated and not changed between Sandwich Shop and the Gowich. |
| GO1_13 | Absolute path to the runtime’s executable. |
| TENANTS | Absolute path to tenants directory. |

### Running

```bash
gowich [--env=<path>] [--port=<port>]
```

| Flag | Description |
| --- | --- |
| --env | Path to a local, or remote .env this process should use for default configurations. |
| --port | The port number to run on. |

### Nodewich

A `Nodewich` is a Sandwich written in Node. You may have as many of these running as needed as long as they are all running on separate ports. Each `Nodewich` should have a `.env` file supplied to it as a CLI argument.

```bash
#.nodewich1.env

PORT=4001
TIMEOUT=60
JWT_SECRET=1234567890jwtsecretcode
NODE9_10=/path/to/node9
NODE12_7=/path/to/node12
TENANTS=/path/to/tenants
```

| Key | Description |
| --- | --- |
| PORT | Port to run this Sandwich on. This will be overridden if --port is supplied as a command argument. |
| TIMEOUT | Timeout, in seconds, to wait for an order to finish. |
| JWT_SECRET | Must be the same JWT_SECRET as specified in the Sandwich Shop’s .env variables. This is used to ensure the request was generated and not changed between Sandwich Shop and the Nodewich. |
| NODE9_10 | Absolute path to the runtime’s executable. |
| TENANTS | Absolute path to tenants directory. |

### Running

```bash
nodewich [--env=<path>] [--port=<port>]
```

| Flag | Description |
| --- | --- |
| --env | Path to a local, or remote .env this process should use for default configurations. |
| --port | The port number to run on. |

### Tenants

Tenants are the buckets where your functional code is placed. Within the `tenant` directory, you will place the different tenant directories. Along with the functions a tenant wishes to run, there also must be an `orders.yml` file to hold the function configurations. There can also either be an optional `.key` file to hold the API key for this particular tenant, or their API key must be stored within the `tenants` collection on your `sandwich-shop` MongoDB database.

```bash
tenants
└── b78682b3-36c8-4759-b8d1-5e62f029a1bc
    ├── .key (optional)
    ├── getSandwich.js
    ├── make_sandwich.go
    └── orders.yml
```

```yaml
# order.yml
---
getSandwich: # order name, GET https://shop.example/{tenantid}/getSandwich  
  runtime: node9_10 # runtime needed, used to pick sandwich 
  path: getSandwich.js # path function is relative to tenant's root  
  env: [] # any variables to include
makeSandwich:  
  runtime: go1_13
  path: make_sandwich.go
  env: []
```

`.key` file

```
this-is-my-api-key
```

**OR**

`sandwich-shop.tenants` in your MongoDB cluster

```json
{
  "_id": "tenant id",
  "key": "this-is-my-api-key"
}
```

## TL;DR - `sudo make me a gowich`

```bash
mkdir ~/sandwich-shop
go install github.com/suhay/sandwich-shop@latest
go install github.com/suhay/sandwich-shop/sandwiches/gowich@latest

cd ~/sandwich-shop
printf "JWT_SECRET=1234567890jwtsecretcode\nTENANTS=~/sandwich-shop" > .env
printf "JWT_SECRET=1234567890jwtsecretcode\nGO1_17=/usr/local/go/bin/go" > .node1.env
echo '[{"_id": "ciabatta", "name": "Ciabatta", "host": "http://127.0.0.1", "port": 4007, "runtimes": ["go1_17"]}]' > sandwiches.json
mkdir myFunctions

cd myFunctions
echo "password1234" > .key
printf -- '---\nhello:\n  runtime: go1_17\n  path: hello.go' > orders.yml
printf 'package main\nimport "fmt"\nfunc main() {\nfmt.Printf("%%s", "Hello!")\n}' > hello.go
```

### Run

```
sandwich-shop --env ~/sandwich-shop/.env &
gowich --env ~/sandwich-shop/.node1.env --port 4007 &
```

### Use

```bash
curl -X POST --header "Authorization: Bearer password1234" http://localhost:3002/shop/myFunctions/hello
```

## Current Development

- [x]  ~~Move tenant path into an environment variable~~
- [ ]  CLI for adding `tenants`, standing up and tearing down Sandwiches, and changing `.key` values.
- [ ]  A+B testing against a Serverless pattern (to see if this is even a thing or just something for fun)
- [ ]  Security audit on tenants

### Developing for Sandwich Shop

```
make dev
```