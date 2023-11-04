import { Request, Response } from 'express'
import AuthenticationBodySchema from '../schemas/routes/body/authentication'
import Log from '../services/logging'
import TokenService from '../services/token'

const authenticateRoute = async (req: Request, res: Response) => {
  const authenticationBody = req.body
  const validationResult = await AuthenticationBodySchema.safeParseAsync(authenticationBody)

  if (!validationResult.success) {
    Log.warn(`Invalid authentication request: ${validationResult.error.message}`)

    return res.status(400).json({
      success: false,
      message: 'Invalid credentials'
    })
  }

  const jwt = TokenService.generate(validationResult.data.service)
  return res.status(200).json({
    success: true,
    message: 'Authentication successful',
    data: {
      jwt
    }
  })
}

export default authenticateRoute
