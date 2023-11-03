import express, { type Application } from 'express'
import cors from '../middleware/cors'

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
    HttpService.instance.listen(3000, () => console.log('Server is running on port 3000'))
  }
}
