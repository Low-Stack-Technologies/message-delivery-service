import type { NextFunction, Request, Response } from 'express'

const headersMiddleware = (requiredHeaders: { [key: string]: string }) => {
  return (req: Request, res: Response, next: NextFunction) => {
    const missingHeaders = Object.entries(requiredHeaders)
      .filter(([header, value]) => req.headers[header] !== value)
      .map(([header]) => header)

    if (missingHeaders.length > 0) {
      return res.status(400).json({
        error: 'Missing required headers',
        missingHeaders
      })
    }

    next()
  }
}

export default headersMiddleware
