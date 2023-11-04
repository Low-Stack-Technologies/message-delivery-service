import 'dotenv/config'
import ConfigurationService from './services/configuration'
import EmailService from './services/email'
import HttpService from './services/http'
import Log from './services/logging'
import TokenService from './services/token'

Log.initialize()
await ConfigurationService.load()
TokenService.initialize()
EmailService.initialize()
HttpService.initialize()
