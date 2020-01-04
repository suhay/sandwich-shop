require(`dotenv`).config()

const jwt = require('express-jwt')
const express = require(`express`)
const timeout = require(`connect-timeout`)
const cp = require('child_process')
const bodyParser = require('body-parser')
const logger = require(`morgan`)
const path = require(`path`)

const app = express()
const defaultPort = "4006"

const port = process.env.PORT || defaultPort

app.use(logger(`common`))
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
app.use(bodyParser.json())
app.use(bodyParser.urlencoded({ extended: true }))

app.post('/:tenantID/:order', 
  timeout(`${process.env.TIMEOUT}s`), 
  (req, res) => {
    if (req.user.tenant === req.params.tenantID && req.user.authorized) {
      const child = cp.exec(`${process.env[req.user.runtime.toUpperCase()]} ${req.params.order} '${JSON.stringify(req.body)}'`, 
        {
          cwd: path.resolve(`${process.env.TENANTS || `../tenants`}/${req.params.tenantID}`),
          timeout: process.env.TIMEOUT * 1000
        },
        (err, stdout, stderr) => {
          if (err) {
            console.error(err)
            res.status(500).send('There was an error')
          } else if (stderr) {
            console.error(stderr)
            res.status(500).send('There was an error')
          } else {
            res.status(200).type('application/json').send(stdout)
          }
        }
      )
    } else {
      res.status(401).send('Not Authorized')
    }
  }
)

app.listen(port, () => console.log(`Nodewich online: http://localhost:${port}/`))
