import type { NextFunction, Request, Response } from 'express'

let API_KEYS: string[] = []

// Get the API keys from the environment variable.
const loadApiKeys = () => {
  API_KEYS = (process.env.API_KEYS?.split(',') || []).map((key) => key.split(':')[1])

  // Check if there are no API keys.
  if (API_KEYS.length === 0) console.warn('No API keys found. Set the API_KEYS environment variable.')

  // Check that all API keys are unique.
  if (API_KEYS.length !== new Set(API_KEYS).size) {
    console.error('All API keys must be unique.')
    process.exit(1)
  }

  console.log('API keys loaded.')
}

setTimeout(loadApiKeys, 500)

/**
 * Middleware to check if the request has a valid API key.
 *
 * @param req - The request.
 * @param res - The response.
 * @param next - The next middleware.
 */
const checkApiKey = (req: Request, res: Response, next: NextFunction) => {
  // Get the API key from the request.
  const apiKey = req.headers['x-api-key'] as string

  // Check if the API key is valid.
  if (API_KEYS.includes(apiKey)) next()
  else res.status(401).json({ message: 'Invalid API key.' })
}

export default checkApiKey
