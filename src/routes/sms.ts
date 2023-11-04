import type { Request, Response } from 'express'
import SmsBodySchema, { SmsBody } from '../schemas/routes/body/sms'
import Log from '../services/logging'
import SmsService from '../services/sms'

const smsRoute = async (req: Request, res: Response) => {
  const smsBody = req.body as SmsBody
  const validationResult = SmsBodySchema.safeParse(smsBody)

  if (!validationResult.success) {
    Log.warn(`Invalid sms request: ${validationResult.error.message}`)

    return res.status(400).json({
      success: false,
      message: 'Invalid body'
    })
  }

  if (smsBody.useTemplate) await sendWithTemplate(smsBody)
  else await sendWithoutTemplate(smsBody)

  res.status(200).json({
    success: true,
    message: 'Sms sent'
  })
}

const sendWithoutTemplate = async (smsBody: SmsBody) => {
  return SmsService.sendSms(smsBody.from, smsBody.to, smsBody.body!)
}

const sendWithTemplate = async (smsBody: SmsBody) => {
  const templateBody = await SmsService.getTemplateBody(smsBody.template!.name, smsBody.service, smsBody.template!.data)

  return SmsService.sendSms(smsBody.from, smsBody.to, templateBody)
}

export default smsRoute
