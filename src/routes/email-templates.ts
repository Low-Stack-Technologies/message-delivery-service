import type { Request, Response } from 'express'
import { readFile } from 'fs/promises'
import { sendEmail } from './../services/email.js'

/**
 * Send email template.
 *
 * @param req Request object.
 * @param res Response object.
 */
const sendEmailTemplate = async (req: Request, res: Response) => {
  const { templateName } = req.params
  const { to, from, subject: templateSubject, parameters } = req.body

  // Get template content.
  const templateContent = await getTemplate(templateName)
  if (!templateContent) return res.status(404).json({ error: 'Template not found.' })

  // Replace template parameters.
  const subject = replaceParameters(templateSubject, parameters)
  const content = replaceParameters(templateContent, parameters)

  // Send email.
  sendEmail(to, from, subject, content)
    .then(() => res.status(200).json({ success: true }))
    .catch((error) => res.status(500).json({ error: error.message }))
}

/**
 * Get a template by name.
 *
 * @param name Name of the template.
 * @returns Template content or null if not found.
 */
const getTemplate = async (name: string): Promise<string | null> => {
  try {
    const dataPath = process.env.DATA_PATH || './data'
    const templatePath = `${dataPath}/templates/${name}`
    return await readFile(templatePath, 'utf8')
  } catch (error) {
    console.error(error)
    return null
  }
}

/**
 * Replace template and subject parameters.
 *
 * @param input Input string.
 * @param parameters Parameters to replace.
 * @returns String with parameters replaced.
 */
const replaceParameters = (input: string, parameters: Record<string, string>): string => {
  let output = input
  for (const [key, value] of Object.entries(parameters)) {
    output = output.replaceAll(`[{${key}}]`, value)
  }
  return output
}

export default sendEmailTemplate
