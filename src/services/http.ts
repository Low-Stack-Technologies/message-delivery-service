import express, { type Application } from 'express'
import authorizedMiddleware from '../middleware/authorized'
import cors from '../middleware/cors'
import headersMiddleware from '../middleware/headers'
import authenticateRoute from '../routes/authenticate'
import emailRoute from '../routes/email'
import ConfigurationService from './configuration'
import Log from './logging'

export default class HttpService {
  private static instance: Application

  public static initialize() {
    HttpService.instance = express()

    HttpService.registerMiddlewares()
    HttpService.registerRoutes()

    HttpService.start()
  }

  private static registerMiddlewares() {
    HttpService.instance.use(express.json())
    HttpService.instance.use(cors)
  }

  private static registerRoutes() {
    HttpService.instance.post('/v2/authenticate', headersMiddleware({ 'content-type': 'application/json' }), authenticateRoute)
    HttpService.instance.post('/v2/email', authorizedMiddleware, emailRoute)
  }

  private static start() {
    const port = ConfigurationService.get().http.port
    HttpService.instance.listen(port, () => {
      Log.info(`Server started on port ${port}`)
    })
  }
}
