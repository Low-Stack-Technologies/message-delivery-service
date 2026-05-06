/* eslint-disable react-refresh/only-export-components */
import {
  createContext,
  useContext,
  useReducer,
  type Dispatch,
  type ReactNode,
} from 'react'

export type ServiceScope = 'all' | 'email' | 'sms'
export type ServiceStatus = 'active' | 'paused'
export type EmailAccountStatus = 'healthy' | 'warning' | 'offline'
export type ActivityTone = 'info' | 'success' | 'warning' | 'danger'

export type ServiceRecord = {
  id: string
  name: string
  owner: string
  scope: ServiceScope
  status: ServiceStatus
  publicKey: string
  notes: string
  createdAt: string
  lastRerollAt: string | null
}

export type EmailAccountRecord = {
  id: string
  address: string
  displayName: string
  smtpHost: string
  smtpPort: number
  username: string
  password: string
  isDefault: boolean
  status: EmailAccountStatus
  lastTestedAt: string | null
}

export type SmsCredentialRecord = {
  username: string
  password: string
  status: 'connected' | 'stale'
  lastSyncedAt: string
  rotationCount: number
}

export type SentMessageRecord = {
  id: string
  channel: 'email' | 'sms'
  serviceId: string
  recipients: string[]
  sender: string
  subject: string | null
  contentMode: 'plain' | 'html' | 'template'
  body: string
  templateName: string | null
  createdAt: string
  status: 'queued' | 'accepted'
}

export type ActivityRecord = {
  id: string
  title: string
  detail: string
  tone: ActivityTone
  createdAt: string
}

type AdminState = {
  services: ServiceRecord[]
  emailAccounts: EmailAccountRecord[]
  smsCredentials: SmsCredentialRecord
  messages: SentMessageRecord[]
  activity: ActivityRecord[]
}

type AdminStore = {
  state: AdminState
  dispatch: Dispatch<Action>
}

type Action =
  | {
      type: 'service/add'
      payload: Omit<ServiceRecord, 'createdAt' | 'lastRerollAt'>
    }
  | {
      type: 'service/delete'
      payload: { id: string }
    }
  | {
      type: 'service/reroll'
      payload: { id: string }
    }
  | {
      type: 'service/set-status'
      payload: { id: string; status: ServiceStatus }
    }
  | {
      type: 'service/set-scope'
      payload: { id: string; scope: ServiceScope }
    }
  | {
      type: 'email/add'
      payload: Omit<EmailAccountRecord, 'isDefault' | 'status' | 'lastTestedAt'>
    }
  | {
      type: 'email/delete'
      payload: { id: string }
    }
  | {
      type: 'email/set-default'
      payload: { id: string }
    }
  | {
      type: 'email/test'
      payload: { id: string }
    }
  | {
      type: 'sms/update'
      payload: Pick<SmsCredentialRecord, 'username' | 'password'>
    }
  | {
      type: 'sms/rotate'
    }
  | {
      type: 'message/send'
      payload: Omit<SentMessageRecord, 'id' | 'createdAt' | 'status'>
    }

const initialState: AdminState = {
  services: [
    {
      id: 'billing-api',
      name: 'Billing API',
      owner: 'Core Platform',
      scope: 'all',
      status: 'active',
      publicKey:
        'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBillingApiExampleKey',
      notes: 'Primary transactional consumer for invoices and receipts.',
      createdAt: '2026-04-29T08:10:00.000Z',
      lastRerollAt: '2026-05-01T10:35:00.000Z',
    },
    {
      id: 'support-hub',
      name: 'Support Hub',
      owner: 'Customer Care',
      scope: 'email',
      status: 'paused',
      publicKey:
        'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAISupportHubExampleKey',
      notes: 'Used for ticket replies and customer follow-ups.',
      createdAt: '2026-04-21T13:25:00.000Z',
      lastRerollAt: null,
    },
    {
      id: 'alerts-worker',
      name: 'Alerts Worker',
      owner: 'Platform Automation',
      scope: 'sms',
      status: 'active',
      publicKey:
        'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAlertsWorkerExampleKey',
      notes: 'Nightly alerting and incident notifications.',
      createdAt: '2026-04-18T17:15:00.000Z',
      lastRerollAt: '2026-04-30T07:40:00.000Z',
    },
  ],
  emailAccounts: [
    {
      id: 'support',
      address: 'support@example.com',
      displayName: 'Support Desk',
      smtpHost: 'smtp.mail.example.com',
      smtpPort: 587,
      username: 'support@example.com',
      password: '••••••••••••',
      isDefault: true,
      status: 'healthy',
      lastTestedAt: '2026-05-04T09:20:00.000Z',
    },
    {
      id: 'receipts',
      address: 'receipts@example.com',
      displayName: 'Receipts',
      smtpHost: 'smtp.send.example.com',
      smtpPort: 465,
      username: 'receipts@example.com',
      password: '••••••••••••',
      isDefault: false,
      status: 'warning',
      lastTestedAt: '2026-05-03T14:45:00.000Z',
    },
  ],
  smsCredentials: {
    username: 'api_user_id',
    password: '••••••••••••',
    status: 'connected',
    lastSyncedAt: '2026-05-04T16:10:00.000Z',
    rotationCount: 2,
  },
  messages: [
    {
      id: 'msg-1001',
      channel: 'email',
      serviceId: 'billing-api',
      recipients: ['finance@example.com'],
      sender: 'support@example.com',
      subject: 'Invoice ready',
      contentMode: 'plain',
      body: 'Your invoice is ready for download.',
      templateName: null,
      createdAt: '2026-05-04T15:20:00.000Z',
      status: 'accepted',
    },
    {
      id: 'msg-1002',
      channel: 'sms',
      serviceId: 'alerts-worker',
      recipients: ['+46700000000'],
      sender: 'AlertOps',
      subject: null,
      contentMode: 'template',
      body: 'Template payload for incident notification',
      templateName: 'incident-sms',
      createdAt: '2026-05-04T18:40:00.000Z',
      status: 'queued',
    },
  ],
  activity: [
    {
      id: 'activity-1',
      title: 'Service key rotated',
      detail: 'Billing API generated a new signing key.',
      tone: 'success',
      createdAt: '2026-05-04T10:35:00.000Z',
    },
    {
      id: 'activity-2',
      title: 'SMTP warning',
      detail: 'Receipts account needs a reconnect check.',
      tone: 'warning',
      createdAt: '2026-05-03T14:45:00.000Z',
    },
    {
      id: 'activity-3',
      title: 'SMS credentials synced',
      detail: '46elks credentials were validated successfully.',
      tone: 'info',
      createdAt: '2026-05-04T16:10:00.000Z',
    },
  ],
}

const AdminStoreContext = createContext<AdminStore | null>(null)

function nowIso() {
  return new Date().toISOString()
}

function makeId(prefix: string) {
  return `${prefix}-${crypto.randomUUID()}`
}

function makeFakeKey(label: string) {
  const token = btoa(`${label}-${crypto.randomUUID()}`)
  return `ssh-ed25519 ${token.slice(0, 44)}`
}

function appendActivity(
  state: AdminState,
  activity: Omit<ActivityRecord, 'id' | 'createdAt'>,
) {
  const entry: ActivityRecord = {
    id: makeId('activity'),
    createdAt: nowIso(),
    ...activity,
  }

  return [entry, ...state.activity].slice(0, 12)
}

function reducer(state: AdminState, action: Action): AdminState {
  switch (action.type) {
    case 'service/add': {
      const service: ServiceRecord = {
        ...action.payload,
        createdAt: nowIso(),
        lastRerollAt: null,
      }

      return {
        ...state,
        services: [service, ...state.services],
        activity: appendActivity(state, {
          title: 'Service added',
          detail: `${service.name} was added to the access registry.`,
          tone: 'success',
        }),
      }
    }
    case 'service/delete': {
      const target = state.services.find((service) => service.id === action.payload.id)
      if (!target) {
        return state
      }

      return {
        ...state,
        services: state.services.filter((service) => service.id !== action.payload.id),
        activity: appendActivity(state, {
          title: 'Service deleted',
          detail: `${target.name} was removed from the registry.`,
          tone: 'danger',
        }),
      }
    }
    case 'service/reroll': {
      const target = state.services.find((service) => service.id === action.payload.id)
      if (!target) {
        return state
      }

      return {
        ...state,
        services: state.services.map((service) =>
          service.id === action.payload.id
            ? {
                ...service,
                publicKey: makeFakeKey(service.id),
                lastRerollAt: nowIso(),
              }
            : service,
        ),
        activity: appendActivity(state, {
          title: 'Key rerolled',
          detail: `${target.name} received a fresh signing key.`,
          tone: 'success',
        }),
      }
    }
    case 'service/set-status': {
      return {
        ...state,
        services: state.services.map((service) =>
          service.id === action.payload.id
            ? { ...service, status: action.payload.status }
            : service,
        ),
        activity: appendActivity(state, {
          title: 'Service access updated',
          detail: `Service ${action.payload.id} is now ${action.payload.status}.`,
          tone: 'info',
        }),
      }
    }
    case 'service/set-scope': {
      return {
        ...state,
        services: state.services.map((service) =>
          service.id === action.payload.id
            ? { ...service, scope: action.payload.scope }
            : service,
        ),
        activity: appendActivity(state, {
          title: 'Scope changed',
          detail: `Service ${action.payload.id} can now access ${action.payload.scope}.`,
          tone: 'info',
        }),
      }
    }
    case 'email/add': {
      const account: EmailAccountRecord = {
        ...action.payload,
        isDefault: state.emailAccounts.length === 0,
        status: 'healthy',
        lastTestedAt: nowIso(),
      }

      return {
        ...state,
        emailAccounts: [account, ...state.emailAccounts],
        activity: appendActivity(state, {
          title: 'Email account added',
          detail: `${account.address} is now available for delivery.`,
          tone: 'success',
        }),
      }
    }
    case 'email/delete': {
      const target = state.emailAccounts.find(
        (account) => account.id === action.payload.id,
      )
      if (!target) {
        return state
      }

      const remaining = state.emailAccounts.filter(
        (account) => account.id !== action.payload.id,
      )
      const defaultId = target.isDefault ? remaining[0]?.id ?? null : state.emailAccounts.find(
        (account) => account.isDefault && account.id !== target.id,
      )?.id ?? null
      const nextState = {
        ...state,
        emailAccounts: remaining.map((account) => ({
          ...account,
          isDefault: account.id === defaultId,
        })),
      }

      return {
        ...nextState,
        activity: appendActivity(state, {
          title: 'Email account deleted',
          detail: `${target.address} was removed from SMTP configuration.`,
          tone: 'danger',
        }),
      }
    }
    case 'email/set-default': {
      return {
        ...state,
        emailAccounts: state.emailAccounts.map((account) => ({
          ...account,
          isDefault: account.id === action.payload.id,
        })),
        activity: appendActivity(state, {
          title: 'Default email account changed',
          detail: `Account ${action.payload.id} is now the default sender.`,
          tone: 'info',
        }),
      }
    }
    case 'email/test': {
      return {
        ...state,
        emailAccounts: state.emailAccounts.map((account) =>
          account.id === action.payload.id
            ? {
                ...account,
                status: 'healthy',
                lastTestedAt: nowIso(),
              }
            : account,
        ),
        activity: appendActivity(state, {
          title: 'SMTP test succeeded',
          detail: `Account ${action.payload.id} passed a local connection check.`,
          tone: 'success',
        }),
      }
    }
    case 'sms/update': {
      return {
        ...state,
        smsCredentials: {
          ...state.smsCredentials,
          ...action.payload,
          lastSyncedAt: nowIso(),
          status: 'connected',
        },
        activity: appendActivity(state, {
          title: '46elks credentials updated',
          detail: 'The SMS credential pair was replaced in the local UI.',
          tone: 'success',
        }),
      }
    }
    case 'sms/rotate': {
      return {
        ...state,
        smsCredentials: {
          ...state.smsCredentials,
          password: '••••••••••••',
          lastSyncedAt: nowIso(),
          rotationCount: state.smsCredentials.rotationCount + 1,
          status: 'connected',
        },
        activity: appendActivity(state, {
          title: '46elks credentials rotated',
          detail: 'A new mock password was generated for the provider account.',
          tone: 'warning',
        }),
      }
    }
    case 'message/send': {
      return {
        ...state,
        messages: [
          {
            id: makeId('msg'),
            createdAt: nowIso(),
            status: 'accepted',
            ...action.payload,
          },
          ...state.messages,
        ],
        activity: appendActivity(state, {
          title: 'Message queued',
          detail: `${action.payload.channel.toUpperCase()} delivery prepared for ${action.payload.recipients.length} recipient(s).`,
          tone: 'success',
        }),
      }
    }
    default:
      return state
  }
}

export function AdminStoreProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(reducer, initialState)

  return (
    <AdminStoreContext.Provider value={{ state, dispatch }}>
      {children}
    </AdminStoreContext.Provider>
  )
}

export function useAdminStore() {
  const value = useContext(AdminStoreContext)

  if (!value) {
    throw new Error('useAdminStore must be used inside AdminStoreProvider')
  }

  return value
}
