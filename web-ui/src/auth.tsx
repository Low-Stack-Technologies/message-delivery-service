/* eslint-disable react-refresh/only-export-components */
import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import {
  adminLogin,
  adminMe,
  type AdminLoginResponse,
  type AdminUser,
} from './admin-api'

const STORAGE_KEY = 'mds-admin-token'

type AdminAuthState = {
  status: 'loading' | 'authenticated' | 'unauthenticated'
  token: string | null
  user: AdminUser | null
  error: string | null
  login: (input: {
    username: string
    password: string
    totpCode: string
  }) => Promise<AdminLoginResponse>
  logout: () => void
}

const AdminAuthContext = createContext<AdminAuthState | null>(null)

export function AdminAuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem(STORAGE_KEY))
  const [user, setUser] = useState<AdminUser | null>(null)
  const [status, setStatus] = useState<AdminAuthState['status']>(
    token ? 'loading' : 'unauthenticated',
  )
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    async function restoreSession() {
      if (!token) {
        setUser(null)
        setStatus('unauthenticated')
        return
      }

      setStatus('loading')
      try {
        const response = await adminMe(token)
        if (cancelled) {
          return
        }
        setUser(response.user)
        setStatus('authenticated')
        setError(null)
      } catch (err) {
        if (cancelled) {
          return
        }
        localStorage.removeItem(STORAGE_KEY)
        setToken(null)
        setUser(null)
        setStatus('unauthenticated')
        setError(err instanceof Error ? err.message : 'Session expired')
      }
    }

    void restoreSession()

    return () => {
      cancelled = true
    }
  }, [token])

  const value = useMemo<AdminAuthState>(
    () => ({
      status,
      token,
      user,
      error,
      login: async (input) => {
        const response = await adminLogin(input)
        localStorage.setItem(STORAGE_KEY, response.token)
        setToken(response.token)
        setUser(response.user)
        setStatus('authenticated')
        setError(null)
        return response
      },
      logout: () => {
        localStorage.removeItem(STORAGE_KEY)
        setToken(null)
        setUser(null)
        setStatus('unauthenticated')
        setError(null)
      },
    }),
    [error, status, token, user],
  )

  return <AdminAuthContext.Provider value={value}>{children}</AdminAuthContext.Provider>
}

export function useAdminAuth() {
  const value = useContext(AdminAuthContext)
  if (!value) {
    throw new Error('useAdminAuth must be used within AdminAuthProvider')
  }
  return value
}
