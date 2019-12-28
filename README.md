# Sandwich Shop
## Serverless experiment using Go and GraphQL

*Note! This is purely an experiment and not meant for production use.*

The purpose of this experiment is to explore other alternatives to the container based serverless pattern. We are hoping this solution will minimize the overhead needed for spinning up a new container on demand and fully eliminate the idea of a "cold start".

More to come soon!

## Usage

### Install

The current version of Sandwich Shop has a prepackaged Go and Node based sandwich it can make. To install the main shop, from the project root, type:

```bash
make install
```

To add the Go and Node sandwiches to the menu, from the project root, type:

```bash
make sandwiches
```

You will need to provide your own version of Node and Go in order to run these modules.

## Setting up shop

Once you have the shop installed, and any of the sandwiches added that you need, the next task is to set up your `.env` files. These can live anywhere and are included as arguments when launching the shop or sandwiches.

- MONGODB_URL: URL to the mongo database that stores tenant information including the `Bearer` token.
- MONGODB_USER: User for logging into the tenant info database.
- MONGODB_PASSWD: Password for the user above.
- JWT_SECRET: Used for signing a JWT payload while passing it through the shop as an order. This is used to verify the order came through a shop directly and that it was unchanged while in transit. This secret must be identical across all shops and sandwich makers.
- TIMEOUT: Seconds to wait on the order to complete.

### Tenant

More soon!

### Gowich

More soon!

### Nodewich

More soon!