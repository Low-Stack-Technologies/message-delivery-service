import { readFile } from 'fs/promises'
import { parsePhoneNumber, type CountryCode } from 'libphonenumber-js'
import path from 'path'
import fileExists from '../libs/fileExists'
import fillTemplate from '../libs/fillTemplate'
import SmsResponseSchema from '../schemas/smsResponse'
import ConfigurationService from './configuration'
import EmailService from './email'
import Log from './logging'

export default class SmsService {
  public static parsePhoneNumber(phoneNumber: string, country: CountryCode): string | false {
    try {
      const parsedPhoneNumber = parsePhoneNumber(phoneNumber, 'SE')
      if (!parsedPhoneNumber) throw new Error('Invalid phone number.')

      return parsedPhoneNumber.formatInternational()
    } catch (error) {
      Log.warn(`Invalid phone number: (${country}) ${phoneNumber}`)
      return false
    }
  }

  public static async sendSms(from: { name: string }, to: { phone: string; country: CountryCode }, body: string): Promise<void> {
    const internationalPhoneNumber = SmsService.parsePhoneNumber(to.phone, to.country)
    if (!internationalPhoneNumber) throw new Error('Invalid phone number.')

    const providerConfig = ConfigurationService.get().sms['46elks']
    const auth = Buffer.from(`${providerConfig.username}:${providerConfig.password}`).toString('base64')

    const requestBody = `from=${encodeURIComponent(from.name)}&to=${encodeURIComponent(internationalPhoneNumber)}&message=${encodeURIComponent(body)}`
    const response = await fetch('https://api.46elks.com/a1/sms', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
        'Authorization': `Basic ${auth}`
      },
      body: requestBody
    })

    if (!response.ok) {
      Log.error(`Failed to send SMS to ${internationalPhoneNumber}. Response: ${response.status} ${response.statusText}`)
      throw new Error('Failed to send SMS.')
    }

    const smsResponse = SmsResponseSchema.safeParse(await response.json())
    if (!smsResponse.success) {
      Log.warn(`Failed to parse SMS response. Response: ${response.status} ${response.statusText}`)
      return
    }

    Log.info(`Sent SMS to ${internationalPhoneNumber}. Response: ${response.status} ${response.statusText}`)
  }

  public static async getTemplateBody(template: string, service: string, data: Record<string, string>): Promise<string> {
    const templateBody = await SmsService.getTemplate(template, service)
    return fillTemplate(templateBody, data)
  }

  private static async getTemplate(template: string, service: string): Promise<string> {
    const templatePath = path.join(EmailService.DATA_PATH, `./templates/sms/${service}/${template}.txt`)
    if (!(await fileExists(templatePath))) throw new Error(`Template ${template} not found`)
    const templateBody = await readFile(templatePath, 'utf-8')

    return templateBody
  }
}
