import { z } from 'zod'

const SmsResponseSchema = z.object({
  status: z.string(),
  direction: z.string(),
  from: z.string(),
  created: z.string(),
  parts: z.number(),
  to: z.string(),
  cost: z.number(),
  message: z.string(),
  id: z.string()
})

export type SmsResponse = z.infer<typeof SmsResponseSchema>

export default SmsResponseSchema
