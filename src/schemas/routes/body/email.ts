import { z } from 'zod'
import ConfigurationService from '../../../services/configuration'

const EmailBodySchema = z
  .object({
    service: z.string(),
    to: z.string().email(),
    from: z.object({
      name: z.string(),
      email: z.string().email()
    }),
    subject: z.string(),

    useTemplate: z.boolean(),

    body: z.string().optional(),
    isHTML: z.boolean().default(true),

    template: z
      .object({
        name: z.string(),
        data: z.object({})
      })
      .optional()
  })
  .refine((data) => {
    if (data.useTemplate && !data.template) return false
    if (!data.useTemplate && !data.body) return false

    return true
  }, 'Invalid body')
  .refine((data) => {
    const services = ConfigurationService.get().services
    const service = services.find((service) => service.details.name === data.service)
    if (!service) return false
    if (!service.access.email.includes(data.from.email)) return false

    return true
  }, 'Invalid service')
  .refine((data) => {
    const emails = ConfigurationService.get().email
    const email = emails.find((email) => email.auth.user === data.from.email)
    if (!email) return false

    return true
  }, 'Invalid email')

export type EmailBody = z.infer<typeof EmailBodySchema>

export default EmailBodySchema
