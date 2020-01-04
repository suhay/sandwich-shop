# Sandwich Shop
## Serverless experiment using Go and GraphQL

*Note! This is purely an experiment and not meant for production use.*

The purpose of this experiment is to explore other alternatives to the container based serverless pattern. We are hoping this solution will minimize the overhead needed for spinning up a new container on demand and fully eliminate the idea of a "cold start".

## Usage

### Install

The current version of Sandwich Shop has a prepackaged Go and Node based sandwich it can make. To install the main shop, from the project root, type:

```bash
$ make install
```

To add the Go and Node sandwiches to the menu, from the project root, type:

```bash
$ make sandwiches
```

You will need to provide your own version of Node and Go in order to run these modules.

### Run

```bash
$ sandwich-shop [--env=<path>] [--port=<port>]
```

|||
|-|-|
|`--env`|Path to a local, or remote `.env` this process should use for default configurations.|
|`--port`|The port number to run on.|

## Setting up shop

Once you have Sandwich Shop installed, and any of the sandwiches added that you need, the next task is to set up your `.env` files. These can live anywhere and are included as arguments when launching the shop or sandwiches.

```
MONGODB_URL=cluster0.mongodb.net
MONGODB_USER=mango
MONGODB_PASSWD=1234password
JWT_SECRET=1234567890jwtsecretcode
TIMEOUT=60
```

|||
|---|---|
|`MONGODB_URL`|URL to the mongo database that stores tenant information including the `Bearer` token.|
|`MONGODB_USER`|User for logging into the tenant info database.|
|`MONGODB_PASSWD`|Password for the user above.|
|`JWT_SECRET`|Used for signing a JWT payload while passing it through the shop as an order. This is used to verify the order came through a shop directly and that it was unchanged while in transit. This secret must be identical across all shops and sandwich makers.|
|`TIMEOUT`|Seconds to wait on the order to complete.|

### Registering a shop

A shop is what is handling the heavy lifting. Each shop is specialized in running one type of application (Node.js, Golang, Python, etc.). There are two ways you can register your shops. The first way is through a local `shops.json` file. This is the more clunky way to maintain your shop registry, but it cuts out a round trip to a MongoDB.

```json
[
  {
    "_id": "rye",
    "name": "Rye",
    "host": "http://localhost:4006",
    "runtimes": [
      "node12_7"
    ]
  },
  {
    "_id": "ciabatta",
    "name": "Ciabatta",
    "host": "http://localhost:4007",
    "runtimes": [
      "go1_13"
    ]
  }
]
```

|||
|---|---|
|`_id`|Shop's id|
|`name`|Human readable name, used more for easily identifying when there is an error (and the shop ids are uuids).|
|`host`|URL and port the shop will be reached from by the main Sandwich Shop process (how it will receive orders).|
|`runtimes`|Array of runtime identifiers that show what this shop is set up to run.|

At the time of writing this README, the current available `runtimes` are `node9_10`,
`node10_6`, `node12_7`, and `go1_13`. These are within a GraphQL `enum` so adding to this list will require a pull request.

The above schema can also be stored within a MongoDB collection called `shops` on the `sandwich-shop` database you are using to run this application.

### Gowich

A `Gowich` is a shop designed to run Golang applications. You may have as many of these running as needed as long as they are all running on separate ports. Each `Gowich` should have a `.env` file, or one supplied to it as a CLI argument.

```
PORT=3001
TIMEOUT=60
JWT_SECRET=1234567890jwtsecretcode
GO1_13=/path/to/go1.13
TENANTS=/path/to/tenants
```

|||
|-|-|
|`PORT`|Port to run this shop on. This will be overridden if `--port` is supplied as a command argument.|
|`TIMEOUT`|Timeout, in seconds, to wait for an order to finish.|
|`JWT_SECRET`|Must be the same `JWT_SECRET` as specified in the main Sandwich Shop's `.env` variables. This is used to ensure the request was generated and not changed between Sandwich Shop and the Gowich.|
|`GO1_13`|Absolute path to the runtime's binary executable.|
|`TENANTS`|Absolute path to tenants directory.|

#### Running

```bash
$ gowich [--env=<path>] [--port=<port>]
```

|||
|-|-|
|`--env`|Path to a local, or remote `.env` this process should use for default configurations.|
|`--port`|The port number to run on.|

### Nodewich

A `Nodewich` is a shop designed to run Node applications. You may have as many of these running as needed as long as they are all running on separate ports. Each `Nodewich` should have a `.env` file, or one supplied to it as a CLI argument.

```
PORT=4001
TIMEOUT=60
JWT_SECRET=1234567890jwtsecretcode
NODE9_10=/path/to/node9
NODE12_7=/path/to/node12
TENANTS=/path/to/tenants
```

|||
|-|-|
|`PORT`|Port to run this shop on. This will be overridden if `--port` is supplied as a command argument.|
|`TIMEOUT`|Timeout, in seconds, to wait for an order to finish.|
|`JWT_SECRET`|Must be the same `JWT_SECRET` as specified in the main Sandwich Shop's `.env` variables. This is used to ensure the request was generated and not changed between Sandwich Shop and the Gowich.|
|`NODE9_10`, `NODE12_7`|Absolute path to the runtime's binary executable.|
|`TENANTS`|Absolute path to tenants directory.|

#### Running

CLI arguments are still in development.

```bash
$ nodewich 
```
<!--[--env=<path>] [--port=<port>]-->
<!--|||
|-|-|
|`--env`|Path to a local, or remote `.env` this process should use for default configurations.|
|`--port`|The port number to run on.|-->

### Tenants

Tenants are the buckets where your functional code is placed. At the writing of this README, the `tenant` directory must be relative to the application root.

Within the `tenant` directory, you will place the different tenant directories. These will eventually be locked down to the individual tenants to prevent directory level access to the shops so tenants will not be able to execute each other's code. Along with the functions a tenant wishes to run, there also must be an `orders.yml` file to hold the function configurations. There should also either be an optional `.key` file to hold the API key for this particular tenant, or their API key must be stored within the `tenants` collection on your `sandwich-shop` MongoDB database.

```
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
getSandwich: # order name, will be called via https://shop.example/{tenantid}/getSandwich
  runtime: node9_10 # shop to run in
  path: getSandwich.js # path where function lives relative to this tenant's root
  env: [] # any variables to include
makeSandwich:
  runtime: go1_13
  path: make_sandwich.go
  env: []
```

*.key* file
```
this-is-my-api-key
```
**OR**

`sandwich-shop.tenants` in your MongoDB cluster
```
{
  "_id": "tenant id",
  "key": "this-is-my-api-key"
}
```

## TL;DR - `sudo make me a gowich`

```bash
$ mkdir sandwich-shop

$ cd sandwich-shop

$ go mod init sandwiches.example.com

$ go get github.com/suhay/sandwich-shop

$ make -C $GOPATH/src/github.com/suhay/sandwich-shop/ install

$ make -C $GOPATH/src/github.com/suhay/sandwich-shop/ gowich

$ touch .env
```

### `.env`

```
MONGODB_URL=cluster0.mongodb.net
MONGODB_USER=mongo
MONGODB_PASSWD=123password
JWT_SECRET=123jwtsecret
TIMEOUT=60
GO1_13=/path/to/go1.13
TENANTS=/path/to/tenants
```

### Run

```bash
$ sandwich-shop &

$ gowich &
```

## Current Development

- [ ] Nodewich CLI

- [x] Move tenant path into an environment variable

- [ ] Dockerize `Gowich` and `Nodewich` shops for ease of adding or removing them as needed

- [ ] `Shop Manager` script for adding and editing `tenants` more easily

- [ ] Allowing `Shop Manager` to also have the ability to stand up and tear down shops based upon a CLI argument

- [ ] A+B testing against a Serverless pattern (to see if this is even a thing, or just something for fun)

- [ ] Security audit on tenants

- [ ] Pywich

### Developing for Sandwich Shop

```bash
$ make dev
```