import { z } from 'zod'

const SmsConfigurationSchema = z.object({
  'provider': z.enum(['46elks']),

  '46elks': z.object({
    username: z.string(),
    password: z.string(),
    currency: z.string()
  })
})

export type SmsConfiguration = z.infer<typeof SmsConfigurationSchema>

export default SmsConfigurationSchema
