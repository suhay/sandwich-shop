require('dotenv').config()

const jwt = require('express-jwt')
const express = require('express')
const timeout = require('connect-timeout')
const util = require('util');
const exec = util.promisify(require('child_process').exec);
const { json, urlencoded } = require('body-parser')
const logger = require('morgan')
const path = require('path');
const e = require('connect-timeout');

const app = express()
const defaultPort = '4006'

const port = process.env.PORT || defaultPort

app.use(logger('common'))
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
app.use(urlencoded({ extended: true }))

app.post('/:tenantID/:order', 
  timeout(`${process.env.TIMEOUT || 60}s`), 
  async (req, res) => {
    if (req.user) {
      console.log('running command...')

      if (req.user.tenant === req.params.tenantID) {
        if (!req.user.authorized) {
          if (req.user.auth) {
            const authResult = await placeOrder(req.user.auth, user)
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
        const result = await placeOrder(order, user, req.body)
  
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

const placeOrder = (order, user, body) => {
  return exec(`${process.env[user.runtime.toUpperCase()]} ${order} '${body ? JSON.stringify(body) : ''}'`, 
  {
    cwd: path.resolve(`${process.env.TENANTS || '../tenants'}/${user.tenant}`),
    timeout: process.env.TIMEOUT * 1000
  })
}

app.listen(port, () => console.log('Nodewich online: http://localhost:${port}/'))
