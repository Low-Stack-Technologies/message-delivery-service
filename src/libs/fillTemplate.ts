const fillTemplate = (template: string, data: Record<string, string>) =>
  template.replace(/{{([^{}]*)}}/g, (_, key) => {
    return data[key] || ''
  })

export default fillTemplate
