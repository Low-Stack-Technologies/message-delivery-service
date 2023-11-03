import { z } from 'zod'

const EmailConfigurationSchema = z.object({
  host: z.string().regex(/^[a-zA-Z0-9.-]+$/),
  port: z.number(),
  secure: z.boolean(),
  auth: z.object({
    user: z.string(),
    pass: z.string()
  })
})

export type EmailConfiguration = z.infer<typeof EmailConfigurationSchema>

export default EmailConfigurationSchema
