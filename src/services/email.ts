import { readFile } from 'fs/promises'
import nodemailer, { type Transporter } from 'nodemailer'
import path from 'path'
import fileExists from '../libs/fileExists'
import fillTemplate from '../libs/fillTemplate'
import ConfigurationService from './configuration'

export default class EmailService {
  private static DATA_PATH = process.env.DATA_PATH || './data/'
  private static transporters = new Map<string, Transporter>()

  public static initialize() {
    const emailAccounts = ConfigurationService.get().email

    for (const emailAccount of emailAccounts) {
      const transporter = nodemailer.createTransport({
        host: emailAccount.host,
        port: emailAccount.port,
        secure: emailAccount.secure,
        auth: {
          user: emailAccount.auth.user,
          pass: emailAccount.auth.pass
        }
      })

      this.transporters.set(emailAccount.auth.user, transporter)
    }
  }

  public static hasEmailAccount(email: string) {
    return this.transporters.has(email)
  }

  public static sendEmail(from: { name: string; email: string }, to: string, subject: string, body: string, isHtml: boolean) {
    const transporter = this.transporters.get(from.email)
    if (!transporter) throw new Error(`No email account found for ${from.email}`)

    return transporter.sendMail({
      from: `${from.name} <${from.email}>`,
      to,
      subject,

      html: isHtml ? body : undefined,
      text: isHtml ? undefined : body
    })
  }

  public static async getTemplateBody(template: string, service: string, data: Record<string, string>) {
    const templateBody = await EmailService.getTemplate(template, service)
    return fillTemplate(templateBody, data)
  }

  private static async getTemplate(template: string, service: string) {
    const templatePath = path.join(EmailService.DATA_PATH, `./templates/email/${service}/${template}.html`)
    if (!(await fileExists(templatePath))) throw new Error(`Template ${template} not found`)
    const templateBody = await readFile(templatePath, 'utf-8')

    return templateBody
  }
}
