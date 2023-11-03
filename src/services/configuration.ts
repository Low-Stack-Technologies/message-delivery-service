import { existsSync } from 'fs'
import { readFile, writeFile } from 'fs/promises'
import ConfigurationSchema, { type Configuration } from '../schemas/Configuration'
import Log from './logging'

export default class ConfigurationService {
  private static CONFIGURATION_FILE_PATH = process.env.CONFIGURATION_FILE_PATH || './config.json'
  private static instance: Configuration | undefined

  public static async load() {
    try {
      if (!this.configurationFileExists()) {
        this.instance = this.createDefaultConfiguration()
        await this.saveConfiguration()
      } else {
        const fileContent = await this.readConfigurationFile()
        const parsedContent = this.parseConfigurationFile(fileContent)
        this.validateConfiguration(parsedContent)
        this.instance = parsedContent
      }
    } catch (error) {
      Log.error(`Failed to load configuration: ${error}`)
      throw error
    }
  }

  public static get(): Configuration {
    if (!this.instance) throw new Error('No configuration loaded')
    return this.instance
  }

  public static async saveConfiguration(): Promise<void> {
    try {
      if (!this.instance) throw new Error('No configuration to save')
      ConfigurationService.validateConfiguration(this.instance)
      await writeFile(this.CONFIGURATION_FILE_PATH, JSON.stringify(this.instance, null, 2))
    } catch (error) {
      Log.error(`Failed to save configuration: ${error}`)
      throw error
    }
  }

  private static configurationFileExists(): boolean {
    return existsSync(this.CONFIGURATION_FILE_PATH)
  }

  private static async readConfigurationFile(): Promise<string> {
    return readFile(this.CONFIGURATION_FILE_PATH, 'utf-8')
  }

  private static parseConfigurationFile(fileContent: string): Configuration {
    return JSON.parse(fileContent)
  }

  private static validateConfiguration(configuration: Configuration): void {
    ConfigurationSchema.parse(configuration)
  }

  private static createDefaultConfiguration(): Configuration {
    return {
      http: {
        port: 3000,

        security: {
          cors: {
            enabled: true,
            origin: '*',
            methods: ['GET', 'POST'],
            headers: ['Content-Type', 'Authorization']
          },

          bruteforce: {
            enabled: false,
            maxAttempts: 5,
            timeWindow: 10000,
            maxDelay: 10000
          }
        }
      },
      services: [
        {
          details: {
            name: 'Service 1',
            password: '$argon2...',
            totpSecret: 'JBSWY...'
          },

          access: {
            email: ['noreply@example.com'],
            sms: true
          },

          whitelist: {
            enabled: true,
            ips: ['::1', '127.0.0.1']
          }
        }
      ],
      email: [
        {
          host: 'smtp.example.com',
          port: 465,
          secure: true,
          auth: {
            user: 'noreply@example.com',
            pass: 'password'
          }
        }
      ],
      sms: {
        'provider': '46elks',
        '46elks': {
          username: '',
          password: ''
        }
      }
    }
  }
}
