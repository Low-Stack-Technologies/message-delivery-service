import type { NextFunction, Request, Response } from 'express'
import Log from '../services/logging'
import TokenService from '../services/token'

const authorizedMiddleware = (req: Request, res: Response, next: NextFunction) => {
  const { authorization } = req.headers
  if (!authorization) return fail(res, 'No authorization header')

  const token = authorization.split(' ')[1]
  if (!token) return fail(res, 'No token provided')

  const service = TokenService.validate(token)
  if (!service) return fail(res, 'Invalid token')

  req.body.service = service
  next()
}

const fail = (res: Response, reason: string) => {
  Log.warn(`Unauthorized request: ${reason}`)
  res.status(401).json({ message: 'Unauthorized' })
}

export default authorizedMiddleware
