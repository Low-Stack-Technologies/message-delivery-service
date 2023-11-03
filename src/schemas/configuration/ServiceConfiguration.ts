import { z } from 'zod'

const ServiceConfigurationSchema = z.object({
  details: z.object({
    name: z.string(),
    password: z.string(),
    totpSecret: z.string()
  }),

  access: z.object({
    email: z.array(z.string()),
    sms: z.boolean()
  }),

  whitelist: z.object({
    enabled: z.boolean(),
    ips: z.array(z.string().ip())
  })
})

export type ServiceConfiguration = z.infer<typeof ServiceConfigurationSchema>

export default ServiceConfigurationSchema
