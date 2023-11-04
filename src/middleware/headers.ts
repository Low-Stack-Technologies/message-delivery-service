import type { NextFunction, Request, Response } from 'express'
import Log from '../services/logging'

const headersMiddleware = (requiredHeaders: { [key: string]: string }) => {
  return (req: Request, res: Response, next: NextFunction) => {
    const missingHeaders = Object.entries(requiredHeaders)
      .filter(([header, value]) => req.headers[header] !== value)
      .map(([header]) => header)

    if (missingHeaders.length > 0) {
      Log.warn(`Missing required headers: ${missingHeaders.join(', ')}`)

      return res.status(400).json({
        error: 'Missing required headers',
        missingHeaders
      })
    }

    next()
  }
}

export default headersMiddleware
