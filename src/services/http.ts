import express from 'express'
import checkApiKey from '../middleware/checkApiKey.js'
import sendEmailTemplate from '../routes/email-templates.js'

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

// Add middleware to check the API key.
app.use(checkApiKey)

// Add POST route to send email template.
app.post('/email/templates/:templateName', sendEmailTemplate)

// Start the server.
const PORT = 3000 || process.env.PORT
app.listen(PORT, () => {
  console.log(`Server started on port ${PORT}`)
})
