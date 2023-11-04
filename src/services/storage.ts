import { mkdir, stat, writeFile } from 'fs/promises'
import path from 'path'
import ConfigurationService from './configuration'

export default class StorageService {
  public static readonly DATA_PATH = process.env.DATA_PATH || './data/'

  public static async initialize(): Promise<void> {
    const directories = [...StorageService.getEmailTemplatePaths(), ...StorageService.getSmsTemplatePaths()]

    for (const directory of directories) {
      await StorageService.createDirectory(directory)
    }
  }

  private static getEmailTemplatePaths() {
    const emailServices = ConfigurationService.get().services.filter((service) => service.access.email.length > 0)
    return emailServices.map((service) => path.join(StorageService.DATA_PATH, `./templates/email/${service.details.name}`))
  }

  private static getSmsTemplatePaths() {
    const smsServices = ConfigurationService.get().services.filter((service) => service.access.sms)
    return smsServices.map((service) => path.join(StorageService.DATA_PATH, `./templates/sms/${service.details.name}`))
  }

  private static async createDirectory(directory: string): Promise<void> {
    const directoryExists = await StorageService.directoryExists(directory)
    if (!directoryExists) {
      await StorageService.createDirectoryRecursive(directory)
      await StorageService.writeExampleTemplate(directory)
    }
  }

  private static async directoryExists(directory: string): Promise<boolean> {
    try {
      const stats = await stat(directory)
      return stats.isDirectory()
    } catch (error) {
      return false
    }
  }

  private static async createDirectoryRecursive(directory: string): Promise<void> {
    await mkdir(directory, { recursive: true })
  }

  private static async writeExampleTemplate(directory: string): Promise<void> {
    const exampleTemplate = 'Hello {{NAME}}!\n\nThis is an example template for the message delivery service.'
    const isEmailTemplate = directory.includes('email')
    const templateName = isEmailTemplate ? 'example.html' : 'example.txt'

    return writeFile(path.join(directory, templateName), exampleTemplate)
  }
}
