import 'dotenv/config'
import HttpService from './services/http'
import Log from './services/logging'

HttpService.initialize()
Log.initialize()
