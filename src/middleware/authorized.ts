import type { NextFunction, Request, Response } from 'express'
import TokenService from '../services/token'

const authorizedMiddleware = (req: Request, res: Response, next: NextFunction) => {
  const { authorization } = req.headers
  if (!authorization) return fail(res)

  const token = authorization.split(' ')[1]
  if (!token) return fail(res)

  const service = TokenService.validate(token)
  if (!service) return fail(res)

  req.body.service = service
  next()
}

const fail = (res: Response) => {
  res.status(401).json({ message: 'Unauthorized' })
}

export default authorizedMiddleware
