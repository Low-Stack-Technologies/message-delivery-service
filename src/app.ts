import 'dotenv/config'
import ConfigurationService from './services/configuration'
import HttpService from './services/http'
import Log from './services/logging'

Log.initialize()
await ConfigurationService.load()
HttpService.initialize()
