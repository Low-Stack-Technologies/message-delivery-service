import { authenticator } from 'otplib'

export default class TotpService {
  public static validate(secret: string, token: string): boolean {
    return authenticator.check(token, secret)
  }
}
