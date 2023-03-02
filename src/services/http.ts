import express from 'express'

// Create a new express application instance.
const app: express.Application = express()

// Add middleware to parse the request body.
app.use(express.json())
app.use(express.urlencoded({ extended: true }))

// Add middleware to handle CORS.
app.use((req, res, next) => {
  res.header('Access-Control-Allow-Origin', '*')
  res.header('Access-Control-Allow-Headers', 'Origin, X-Requested-With, Content-Type, Accept')
  next()
})

// Start the server.
const PORT = 3000 || process.env.PORT
app.listen(PORT, () => {
  console.log(`Server started on port ${PORT}`)
})
