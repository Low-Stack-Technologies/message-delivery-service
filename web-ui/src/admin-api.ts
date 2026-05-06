export type AdminUser = {
  id: string
  username: string
  createdAt: string
  updatedAt: string
  lastLoginAt: string | null
}

export type AdminUserCredentials = {
  password: string
  totpSecret: string
  provisioningUri: string
}

export type AdminLoginResponse = {
  success: boolean
  token: string
  expiresAt: string
  user: AdminUser
}

export type AdminMeResponse = {
  success: boolean
  user: AdminUser
}

export type AdminUserListResponse = {
  success: boolean
  users: AdminUser[]
}

export type AdminUserCreateResponse = {
  success: boolean
  user: AdminUser
  credentials: AdminUserCredentials
}

export type AdminServiceScope = 'all' | 'email' | 'sms'
export type AdminServiceEmailAccessMode = 'all' | 'restricted'
export type AdminServiceStatus = 'active' | 'paused'
export type AdminEmailAccountStatus = 'healthy' | 'warning' | 'offline'
export type AdminMessageChannel = 'email' | 'sms'
export type AdminMessageContentMode = 'plain' | 'html' | 'template'
export type AdminMessageStatus = 'queued' | 'accepted'
export type AdminActivityTone = 'info' | 'success' | 'warning' | 'danger'
export type AdminSmsCredentialsStatus = 'connected' | 'stale'

export type AdminService = {
  id: string
  name: string
  owner: string
  scope: AdminServiceScope
  emailAccessMode: AdminServiceEmailAccessMode
  allowedEmailAccountIds: string[]
  status: AdminServiceStatus
  publicKey: string
  notes: string
  createdAt: string
  lastRerollAt: string | null
}

export type AdminServiceCreateRequest = {
  id: string
  name: string
  owner: string
  scope: AdminServiceScope
  emailAccessMode: AdminServiceEmailAccessMode
  allowedEmailAccountIds: string[]
  status: AdminServiceStatus
  publicKey: string
  notes: string
}

export type AdminServiceUpdateRequest = Partial<Pick<
  AdminService,
  'name' | 'owner' | 'scope' | 'emailAccessMode' | 'allowedEmailAccountIds' | 'status' | 'publicKey' | 'notes'
>>

export type AdminServiceListResponse = {
  success: boolean
  services: AdminService[]
}

export type AdminServiceResponse = {
  success: boolean
  service: AdminService
}

export type AdminServiceRerollResponse = {
  success: boolean
  service: AdminService
  privateKey: string
}

export type AdminEmailAccount = {
  id: string
  address: string
  displayName: string
  smtpHost: string
  smtpPort: number
  username: string
  password: string
  isDefault: boolean
  status: AdminEmailAccountStatus
  lastTestedAt: string | null
}

export type AdminEmailAccountCreateRequest = {
  id: string
  address: string
  displayName: string
  smtpHost: string
  smtpPort: number
  username: string
  password: string
  isDefault?: boolean
}

export type AdminEmailAccountUpdateRequest = Partial<{
  address: string
  displayName: string
  smtpHost: string
  smtpPort: number
  username: string
  password: string
  isDefault: boolean
}>

export type AdminEmailAccountListResponse = {
  success: boolean
  emailAccounts: AdminEmailAccount[]
}

export type AdminEmailAccountResponse = {
  success: boolean
  emailAccount: AdminEmailAccount
}

export type AdminEmailAccountTestResponse = {
  success: boolean
  emailAccount: AdminEmailAccount
  testedAt: string
  message?: string
}

type BackendAdminEmailSmtpConfig = {
  host: string
  port: number
  username: string
  password: string
}

type BackendAdminEmailAccount = {
  id: string
  address: string
  displayName?: string
  smtp: BackendAdminEmailSmtpConfig
  isDefault?: boolean
  status: AdminEmailAccountStatus
  lastTestedAt?: string | null
}

type BackendAdminEmailAccountListResponse = {
  success: boolean
  emailAccounts: BackendAdminEmailAccount[]
}

type BackendAdminEmailAccountResponse = {
  success: boolean
  emailAccount: BackendAdminEmailAccount
}

type BackendAdminEmailAccountTestResponse = {
  success: boolean
  emailAccount: BackendAdminEmailAccount
  testedAt: string
  message?: string
}

export type AdminSmsCredentials = {
  username: string
  password: string
  status: AdminSmsCredentialsStatus
  lastSyncedAt: string
  rotationCount: number
}

export type AdminSmsCredentialsUpdateRequest = {
  username: string
  password: string
}

export type AdminSmsCredentialsResponse = {
  success: boolean
  smsCredentials: AdminSmsCredentials
}

export type AdminSmsCredentialsRotateResponse = {
  success: boolean
  smsCredentials: AdminSmsCredentials
}

export type AdminMessageTemplate = {
  name: string
  data: Record<string, unknown>
}

export type AdminMessageRequest = {
  channel: AdminMessageChannel
  serviceId: string
  recipients: string[]
  from?: string
  senderName?: string
  subject?: string
  contentMode: AdminMessageContentMode
  body?: string
  template?: AdminMessageTemplate
}

export type AdminMessage = {
  id: string
  channel: AdminMessageChannel
  serviceId: string
  recipients: string[]
  sender: string
  subject: string | null
  contentMode: AdminMessageContentMode
  body: string | null
  templateName: string | null
  createdAt: string
  status: AdminMessageStatus
}

export type AdminMessageListResponse = {
  success: boolean
  messages: AdminMessage[]
}

export type AdminMessageSubmitResponse = {
  success: boolean
  message: string
  queuedMessage: AdminMessage
}

export type AdminMessagePreviewResponse = {
  success: boolean
  preview: {
    request: AdminMessageRequest
    rendered: string
    warnings?: string[]
  }
}

export type AdminDashboardResponse = {
  success: boolean
  summary: {
    services: number
    activeServices: number
    emailAccounts: number
    defaultEmailAccount: string | null
    smsStatus: AdminSmsCredentialsStatus
    queuedMessages: number
  }
  recentActivity: AdminActivity[]
  recentMessages: AdminMessage[]
}

export type AdminActivity = {
  id: string
  title: string
  detail: string
  tone: AdminActivityTone
  createdAt: string
}

export type AdminDashboardSummary = AdminDashboardResponse['summary']

export async function adminLogin(input: {
  username: string
  password: string
  totpCode: string
}) {
  return requestJson<AdminLoginResponse>('/v3/admin/auth/login', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export async function adminMe(token: string) {
  return requestJson<AdminMeResponse>('/v3/admin/auth/me', {
    method: 'GET',
    headers: authHeaders(token),
  })
}

export async function listAdminUsers(token: string) {
  return requestJson<AdminUserListResponse>('/v3/admin/users', {
    method: 'GET',
    headers: authHeaders(token),
  })
}

export async function createAdminUser(token: string, username: string) {
  return requestJson<AdminUserCreateResponse>('/v3/admin/users', {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify({ username }),
  })
}

export async function getAdminDashboard(token: string) {
  return requestJson<AdminDashboardResponse>('/v3/admin/dashboard', {
    method: 'GET',
    headers: authHeaders(token),
  })
}

export async function listAdminServices(token: string) {
  return requestJson<AdminServiceListResponse>('/v3/admin/services', {
    method: 'GET',
    headers: authHeaders(token),
  })
}

export async function createAdminService(token: string, input: AdminServiceCreateRequest) {
  return requestJson<AdminServiceResponse>('/v3/admin/services', {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify(input),
  })
}

export async function updateAdminService(
  token: string,
  serviceId: string,
  input: AdminServiceUpdateRequest,
) {
  return requestJson<AdminServiceResponse>(`/v3/admin/services/${encodeURIComponent(serviceId)}`, {
    method: 'PUT',
    headers: authHeaders(token),
    body: JSON.stringify(input),
  })
}

export async function deleteAdminService(token: string, serviceId: string) {
  return requestJson<void>(`/v3/admin/services/${encodeURIComponent(serviceId)}`, {
    method: 'DELETE',
    headers: authHeaders(token),
  })
}

export async function rerollAdminService(token: string, serviceId: string) {
  return requestJson<AdminServiceRerollResponse>(
    `/v3/admin/services/${encodeURIComponent(serviceId)}/reroll`,
    {
      method: 'POST',
      headers: authHeaders(token),
    },
  )
}

export async function listAdminEmailAccounts(token: string) {
  const response = await requestJson<BackendAdminEmailAccountListResponse>(
    '/v3/admin/email-accounts',
    {
      method: 'GET',
      headers: authHeaders(token),
    },
  )

  return {
    success: response.success,
    emailAccounts: response.emailAccounts.map(toFrontendEmailAccount),
  }
}

export async function createAdminEmailAccount(
  token: string,
  input: AdminEmailAccountCreateRequest,
) {
  const response = await requestJson<BackendAdminEmailAccountResponse>('/v3/admin/email-accounts', {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify({
      id: input.id,
      address: input.address,
      displayName: input.displayName,
      isDefault: input.isDefault,
      smtp: {
        host: input.smtpHost,
        port: input.smtpPort,
        username: input.username,
        password: input.password,
      },
    }),
  })

  return {
    success: response.success,
    emailAccount: toFrontendEmailAccount(response.emailAccount),
  }
}

export async function updateAdminEmailAccount(
  token: string,
  accountId: string,
  input: AdminEmailAccountUpdateRequest,
) {
  const response = await requestJson<BackendAdminEmailAccountResponse>(
    `/v3/admin/email-accounts/${encodeURIComponent(accountId)}`,
    {
      method: 'PUT',
      headers: authHeaders(token),
      body: JSON.stringify({
        address: input.address,
        displayName: input.displayName,
        isDefault: input.isDefault,
        smtp:
          input.smtpHost || input.smtpPort || input.username || input.password
            ? {
                host: input.smtpHost,
                port: input.smtpPort,
                username: input.username,
                password: input.password,
              }
            : undefined,
      }),
    },
  )

  return {
    success: response.success,
    emailAccount: toFrontendEmailAccount(response.emailAccount),
  }
}

export async function deleteAdminEmailAccount(token: string, accountId: string) {
  return requestJson<void>(`/v3/admin/email-accounts/${encodeURIComponent(accountId)}`, {
    method: 'DELETE',
    headers: authHeaders(token),
  })
}

export async function testAdminEmailAccount(token: string, accountId: string) {
  const response = await requestJson<BackendAdminEmailAccountTestResponse>(
    `/v3/admin/email-accounts/${encodeURIComponent(accountId)}/test`,
    {
      method: 'POST',
      headers: authHeaders(token),
    },
  )

  return {
    success: response.success,
    emailAccount: toFrontendEmailAccount(response.emailAccount),
    testedAt: response.testedAt,
    message: response.message,
  }
}

export async function getAdminSmsCredentials(token: string) {
  return requestJson<AdminSmsCredentialsResponse>('/v3/admin/sms-credentials', {
    method: 'GET',
    headers: authHeaders(token),
  })
}

export async function updateAdminSmsCredentials(
  token: string,
  input: AdminSmsCredentialsUpdateRequest,
) {
  return requestJson<AdminSmsCredentialsResponse>('/v3/admin/sms-credentials', {
    method: 'PUT',
    headers: authHeaders(token),
    body: JSON.stringify(input),
  })
}

export async function rotateAdminSmsCredentials(token: string) {
  return requestJson<AdminSmsCredentialsRotateResponse>('/v3/admin/sms-credentials/rotate', {
    method: 'POST',
    headers: authHeaders(token),
  })
}

export async function listAdminMessages(
  token: string,
  params?: {
    limit?: number
    channel?: AdminMessageChannel
    serviceId?: string
  },
) {
  const searchParams = new URLSearchParams()
  if (params?.limit) {
    searchParams.set('limit', String(params.limit))
  }
  if (params?.channel) {
    searchParams.set('channel', params.channel)
  }
  if (params?.serviceId) {
    searchParams.set('serviceId', params.serviceId)
  }
  const query = searchParams.toString()

  return requestJson<AdminMessageListResponse>(
    query ? `/v3/admin/messages?${query}` : '/v3/admin/messages',
    {
      method: 'GET',
      headers: authHeaders(token),
    },
  )
}

export async function previewAdminMessage(token: string, input: AdminMessageRequest) {
  return requestJson<AdminMessagePreviewResponse>('/v3/admin/messages/preview', {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify(input),
  })
}

export async function createAdminMessage(token: string, input: AdminMessageRequest) {
  return requestJson<AdminMessageSubmitResponse>('/v3/admin/messages', {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify(input),
  })
}

function toFrontendEmailAccount(account: BackendAdminEmailAccount): AdminEmailAccount {
  return {
    id: account.id,
    address: account.address,
    displayName: account.displayName ?? account.address,
    smtpHost: account.smtp.host,
    smtpPort: account.smtp.port,
    username: account.smtp.username,
    password: account.smtp.password,
    isDefault: Boolean(account.isDefault),
    status: account.status,
    lastTestedAt: account.lastTestedAt ?? null,
  }
}

async function requestJson<T>(
  path: string,
  init: RequestInit & { headers?: Record<string, string> },
) {
  const response = await fetch(path, {
    ...init,
    headers: {
      Accept: 'application/json',
      ...(init.body ? { 'Content-Type': 'application/json' } : {}),
      ...(init.headers ?? {}),
    },
  })

  if (!response.ok) {
    const message = await readErrorMessage(response)
    throw new Error(message)
  }

  if (response.status === 204) {
    return undefined as T
  }

  return (await response.json()) as T
}

function authHeaders(token: string) {
  return {
    Authorization: `Bearer ${token}`,
  }
}

async function readErrorMessage(response: Response) {
  try {
    const payload = (await response.json()) as {
      error?: { message?: string }
      message?: string
    }
    return payload.error?.message ?? payload.message ?? response.statusText
  } catch {
    return response.statusText
  }
}
