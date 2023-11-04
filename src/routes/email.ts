import type { Request, Response } from 'express'
import fillTemplate from '../libs/fillTemplate'
import EmailBodySchema, { EmailBody } from '../schemas/routes/body/email'
import EmailService from '../services/email'
import Log from '../services/logging'

const emailRoute = async (req: Request, res: Response) => {
  const emailBody = req.body as EmailBody
  const validationResult = EmailBodySchema.safeParse(emailBody)

  if (!validationResult.success) {
    Log.warn(`Invalid email request: ${validationResult.error.message}`)

    return res.status(400).json({
      success: false,
      message: 'Invalid body'
    })
  }

  if (emailBody.useTemplate) await sendWithTemplate(emailBody)
  else await sendWithoutTemplate(emailBody)

  res.status(200).json({
    success: true,
    message: 'Email sent'
  })
}

const sendWithoutTemplate = async (emailBody: EmailBody) => {
  return EmailService.sendEmail(
    {
      name: emailBody.from.name,
      email: emailBody.from.email
    },
    emailBody.to,
    emailBody.subject,
    emailBody.body!,
    emailBody.isHTML
  )
}

const sendWithTemplate = async (emailBody: EmailBody) => {
  const subject = fillTemplate(emailBody.subject, emailBody.template!.data)
  const templateBody = await EmailService.getTemplateBody(emailBody.template!.name, emailBody.service, emailBody.template!.data)

  return EmailService.sendEmail(
    {
      name: emailBody.from.name,
      email: emailBody.from.email
    },
    emailBody.to,
    subject,
    templateBody,
    true
  )
}

export default emailRoute
