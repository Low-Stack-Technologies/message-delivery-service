import { NextFunction, Request, Response } from 'express'
import ConfigurationService from '../services/configuration'

const cors = (req: Request, res: Response, next: NextFunction) => {
  const config = ConfigurationService.get().http.security.cors

  if (config.enabled) {
    res.header('Access-Control-Allow-Origin', config.origin)
    res.header('Access-Control-Allow-Methods', config.methods)
    res.header('Access-Control-Allow-Headers', config.headers)
  }

  next()
}

export default cors
