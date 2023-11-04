import type { Request, Response } from 'express'
import EmailBodySchema, { EmailBody } from '../schemas/routes/body/email'
import EmailService from '../services/email'
import Log from '../services/logging'

const emailRoute = async (req: Request, res: Response) => {
  const emailBody = req.body as EmailBody
  const validationResult = EmailBodySchema.safeParse(emailBody)

  if (!validationResult.success) {
    Log.verbose(JSON.stringify(validationResult.error, null, 2))

    return res.status(400).json({
      success: false,
      message: 'Invalid body'
    })
  }

  await EmailService.sendEmail(
    {
      name: emailBody.from.name,
      email: emailBody.from.email
    },
    emailBody.to,
    emailBody.subject,
    getBody(emailBody),
    emailBody.isHTML
  )

  res.status(200).json({
    success: true,
    message: 'Email sent'
  })
}

const getBody = (body: EmailBody) => {
  if (body.useTemplate) {
    return 'Template'
  } else {
    return body.body
  }
}

export default emailRoute
