import z from 'zod'

const HttpConfigurationSchema = z.object({
  port: z.number().default(3000),

  security: z.object({
    cors: z.object({
      enabled: z.boolean().default(true),
      origin: z.string().default('*'),
      methods: z.array(z.string()).default(['GET', 'POST']),
      headers: z.array(z.string()).default(['Content-Type', 'Authorization'])
    }),

    bruteforce: z.object({
      enabled: z.boolean().default(true),
      maxAttempts: z.number().default(5),
      timeWindow: z.number().default(300000),
      maxDelay: z.number().default(10000)
    })
  })
})

export type HttpConfiguration = z.infer<typeof HttpConfigurationSchema>

export default HttpConfigurationSchema
