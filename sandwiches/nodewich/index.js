require('dotenv').config()

const jwt = require('express-jwt')
const express = require('express')
const timeout = require('connect-timeout')
const util = require('util');
const exec = util.promisify(require('child_process').exec);
const {
  json,
  urlencoded
} = require('body-parser')
const path = require('path');

const app = express()
const defaultPort = '4006'

const port = process.env.PORT || defaultPort

app.use(
  jwt({
    secret: process.env.JWT_SECRET,
    credentialsRequired: false,
    getToken: (req) => {
      if (req.headers.token && req.headers.token.length > 0) {
        return req.headers.token
      } else if (req.query && req.query.token) {
        return req.query.token
      }
      return null
    }
  })
)
app.use(json())
app.use(urlencoded({
  extended: true
}))

app.post('/:tenantID/:order',
  timeout(`${process.env.TIMEOUT || 60}s`),
  async (req, res) => {
    if (req.user) {
      console.log('running command...')
      const claims = req.user

      if (claims.tenant === req.params.tenantID) {
        const body = req.body

        if (!claims.authorized) {
          if (claims.auth) {
            const authResult = await placeOrder(claims.auth, claims, body)
            if (authResult.stderr || authResult.stdout !== 'true') {
              console.error(authResult.stderr)
              res.status(401).send('Not Authorized')
            }
          } else {
            console.error(authResult.stderr)
            res.status(401).send('Not Authorized')
          }
        }

        const order = req.params.order
        const result = await placeOrder(order, claims, body)

        if (result.stderr) {
          console.error(result.stderr)
          res.status(500).send('There was an error')
        }

        res.status(200).type('application/json').send(result.stdout)
      } else {
        res.status(401).send('Not Authorized')
      }
    } else {
      res.status(401).send('Not Authorized')
    }
  }
)

const parseEnvVariables = (envVariables) => {
  const parsedVars = envVariables.reduce((acc, value) => {
    const parts = value.split('=')
    acc[parts[0]] = parts[1]
    return acc
  }, {})
}

const placeOrder = (order, claims, body) => {
  const envVariables = claims.Env ? JSON.parse(claims.Env) : []

  return exec(`${process.env[claims.runtime.toUpperCase()]} ${order} '${body ? JSON.stringify(body) : ''}'`, {
    cwd: path.resolve(`${process.env.TENANTS || '../tenants'}/${claims.tenant}`),
    timeout: process.env.TIMEOUT * 1000,
    env: {
      ...process.env,
      ...parseEnvVariables(envVariables)
    }
  })
}

app.listen(port, () => console.log('Nodewich online: http://localhost:${port}/'))