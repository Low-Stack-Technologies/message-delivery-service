import { randomBytes } from 'crypto'
import jwt from 'jsonwebtoken'
import ConfigurationService from './configuration'

export default class TokenService {
  private static JWT_SECRET: string

  public static initialize(): void {
    TokenService.JWT_SECRET = ConfigurationService.get().http.security.token.secret
  }

  public static generate(service: string): string {
    return jwt.sign({ data: { service } }, TokenService.JWT_SECRET, { expiresIn: '1h' })
  }

  public static generateSecret(): string {
    return randomBytes(64).toString('hex')
  }
}
