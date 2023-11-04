import { z } from 'zod'
import ConfigurationService from '../../../services/configuration'
import PasswordService from '../../../services/password'
import TotpService from '../../../services/totp'

const AuthenticationBodySchema = z
  .object({
    service: z.string(),
    password: z.string(),
    totp: z.string()
  })
  .refine(async (data) => {
    const services = ConfigurationService.get().services
    const service = services.find((service) => service.details.name === data.service)
    if (!service) return false

    if (!(await PasswordService.validate(data.password, service.details.password))) return false
    if (!TotpService.validate(service.details.totpSecret, data.totp)) return false

    return true
  }, 'Invalid credentials')

export type AuthenticationBody = z.infer<typeof AuthenticationBodySchema>

export default AuthenticationBodySchema
