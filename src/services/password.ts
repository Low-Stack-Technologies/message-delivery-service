import argon2 from 'argon2'

export default class PasswordService {
  public static async validate(plain: string, hash: string): Promise<boolean> {
    return await argon2.verify(hash, plain)
  }
}
