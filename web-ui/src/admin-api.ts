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

async function requestJson<T>(
  path: string,
  init: RequestInit & { headers?: Record<string, string> },
) {
  const response = await fetch(path, {
    ...init,
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
      ...(init.headers ?? {}),
    },
  })

  if (!response.ok) {
    const message = await readErrorMessage(response)
    throw new Error(message)
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
