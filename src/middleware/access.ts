import type { NextFunction, Request, Response } from 'express'
import Log from '../services/logging'

const accessMiddleware = (req: Request, res: Response, next: NextFunction) => {
  Log.access(`${[req.ip, ...req.ips].join(',')} [${req.method}] ${req.url}`)
  next()
}

export default accessMiddleware
