import validator from 'validator'
import { z } from 'zod'
import ConfigurationService from '../../../services/configuration'
import SmsService from '../../../services/sms'
import CountryCodeSchema from '../../countryCode'

const SmsBodySchema = z
  .object({
    service: z.string(),
    to: z
      .object({
        phone: z.string().refine((phone) => validator.isMobilePhone(phone), 'Number is not a valid phone number'),
        country: CountryCodeSchema
      })
      .refine((data) => {
        if (!SmsService.parsePhoneNumber(data.phone, data.country)) return false
        return true
      }, 'Invalid country code or phone number'),
    from: z.object({
      name: z.string()
    }),

    useTemplate: z.boolean(),

    body: z.string().optional(),

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
    if (!service.access.sms) return false

    return true
  }, 'Invalid service')

export type SmsBody = z.infer<typeof SmsBodySchema>

export default SmsBodySchema
