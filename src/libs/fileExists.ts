import { stat } from 'fs/promises'
import Log from '../services/logging'

const fileExists = async (path: string): Promise<boolean> => {
  try {
    await stat(path)
    return true
  } catch {
    Log.error(`File ${path} does not exist`)
    return false
  }
}

export default fileExists
