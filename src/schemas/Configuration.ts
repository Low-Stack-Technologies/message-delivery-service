import { z } from 'zod'
import EmailConfigurationSchema from './EmailConfiguration'
import HttpConfigurationSchema from './HttpConfiguration'
import ServiceConfigurationSchema from './ServiceConfiguration'
import SmsConfigurationSchema from './SmsConfiguration'

const ConfigurationSchema = z.object({
  http: HttpConfigurationSchema,
  services: z.array(ServiceConfigurationSchema),
  email: z.array(EmailConfigurationSchema),
  sms: SmsConfigurationSchema
})

export type Configuration = z.infer<typeof ConfigurationSchema>

export default ConfigurationSchema
