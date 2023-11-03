import express, { type Application } from 'express'
import cors from '../middleware/cors'
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

  private static registerRoutes() {}

  private static start() {
    const port = process.env.PORT || 3000
    HttpService.instance.listen(port, () => {
      Log.info(`Server started on port ${port}`)
    })
  }
}
