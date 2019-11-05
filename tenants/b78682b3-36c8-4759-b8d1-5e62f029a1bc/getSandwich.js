require(`dotenv`).config()

const MongoClient = require("mongodb").MongoClient
const ObjectId = require("mongodb").ObjectID

const CONNECTION_URL = `mongodb+srv://${process.env.MONGODB_USER}:${process.env.MONGODB_PASSWD}@${process.env.MONGODB_URL}?retryWrites=true`
const DATABASE_NAME = `sandwich-shop`

const args = process.argv.slice(2)
const jsonArgs = JSON.parse(args)

MongoClient.connect(CONNECTION_URL, { useNewUrlParser: true, useUnifiedTopology: true }, (error, client) => {
  if (error) {
    process.stderr.write(error.message)
    throw error
  }
  database = client.db(DATABASE_NAME)
  collection = database.collection(`sandwiches`)

  collection.find({ _id: ObjectId(jsonArgs.id) }).toArray((error, result) => {
    if (error) {
      process.stderr.write(error.message)
    }
    process.stdout.write(JSON.stringify(result))
  })

  client.close()
})
