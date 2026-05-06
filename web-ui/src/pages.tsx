import { useEffect, useState } from 'react'
import type {
  ButtonHTMLAttributes,
  InputHTMLAttributes,
  ReactNode,
  SelectHTMLAttributes,
  TextareaHTMLAttributes,
} from 'react'
import { Link, Outlet, useLocation, useNavigate } from '@tanstack/react-router'
import {
  createAdminMessage,
  createAdminUser,
  listAdminUsers,
  previewAdminMessage,
  type AdminMessageRequest,
} from './admin-api'
import {
  useAdminStore,
  type ActivityTone,
  type EmailAccountRecord,
  type ServiceScope,
  type ServiceStatus,
} from './admin-store'
import { useAdminAuth } from './auth'

type BadgeTone = ActivityTone | 'neutral' | 'success' | 'warning'

const channelLabels = {
  email: 'Email',
  sms: 'SMS',
} as const

function formatDate(value: string | null) {
  if (!value) {
    return 'Never'
  }

  return new Intl.DateTimeFormat('en-GB', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value))
}

function splitValues(value: string) {
  return value
    .split(/[\n,]/)
    .map((item) => item.trim())
    .filter(Boolean)
}

function slugify(value: string) {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

function maskSecret(value: string) {
  if (!value) {
    return '—'
  }

  return '•'.repeat(Math.max(8, Math.min(16, value.length)))
}

function toneClass(tone: BadgeTone) {
  return `badge badge-${tone}`
}

function getDefaultEmailAccount(accounts: EmailAccountRecord[]) {
  return accounts.find((account) => account.isDefault) ?? accounts[0] ?? null
}

export function AdminShell() {
  const { state } = useAdminStore()
  const { status, user, logout } = useAdminAuth()

  if (status !== 'authenticated') {
    return <LoginPage />
  }

  const activeServices = state.services.filter((service) => service.status === 'active')
  const queuedMessages = state.messages.filter((message) => message.status === 'queued').length
  const defaultEmail = getDefaultEmailAccount(state.emailAccounts)
  const latestActivity = state.activity[0]

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="brand">
          <div className="brand-mark">MDS</div>
          <div>
            <p className="eyebrow">Admin console</p>
            <h1>Message Delivery</h1>
          </div>
        </div>

        <nav className="nav">
          <Link
            to="/"
            className="nav-link"
            activeProps={{ className: 'nav-link is-active' }}
            activeOptions={{ exact: true }}
          >
            <span>Overview</span>
            <small>Service health, queue, activity</small>
          </Link>
          <Link
            to="/services"
            className="nav-link"
            activeProps={{ className: 'nav-link is-active' }}
          >
            <span>Services</span>
            <small>Access, keys, scopes</small>
          </Link>
          <Link
            to="/users"
            className="nav-link"
            activeProps={{ className: 'nav-link is-active' }}
          >
            <span>Admin users</span>
            <small>Username, password, and TOTP access</small>
          </Link>
          <Link
            to="/email-accounts"
            className="nav-link"
            activeProps={{ className: 'nav-link is-active' }}
          >
            <span>Email accounts</span>
            <small>SMTP senders and defaults</small>
          </Link>
          <Link
            to="/46elks"
            className="nav-link"
            activeProps={{ className: 'nav-link is-active' }}
          >
            <span>46elks credentials</span>
            <small>SMS provider credentials</small>
          </Link>
          <Link
            to="/send-message"
            className="nav-link"
            activeProps={{ className: 'nav-link is-active' }}
          >
            <span>Send message</span>
            <small>Compose and queue deliveries</small>
          </Link>
        </nav>

        <section className="sidebar-card">
          <p className="eyebrow">Live state</p>
          <strong>{queuedMessages} queued messages</strong>
          <span>
            {activeServices.length} active services ·{' '}
            {state.emailAccounts.length} SMTP accounts
          </span>
          <div className="sidebar-meta">
            <span className={toneClass(defaultEmail ? 'success' : 'warning')}>
              {defaultEmail ? `Default sender: ${defaultEmail.address}` : 'No default sender'}
            </span>
            <span className={toneClass('neutral')}>
              Latest: {latestActivity ? latestActivity.title : 'No activity'}
            </span>
          </div>
        </section>
      </aside>

      <div className="content-shell">
        <header className="topbar">
          <div>
            <p className="eyebrow">Operations surface</p>
            <h2>Manage services, providers, and outbound delivery.</h2>
          </div>
        <div className="topbar-meta">
          <span className="pill">Live admin data</span>
          <span className="pill pill-muted">
            {user ? `Signed in as ${user.username}` : 'Not signed in'}
          </span>
          <span className="pill pill-muted">Updated {formatDate(latestActivity?.createdAt ?? null)}</span>
          <button type="button" className="button button-secondary" onClick={logout}>
            Log out
          </button>
        </div>
      </header>

        <main className="content">
          <Outlet />
        </main>
      </div>
    </div>
  )
}

export function LoginPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const { status, error, login } = useAdminAuth()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [totpCode, setTotpCode] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [message, setMessage] = useState<string | null>(null)

  return (
    <div className="auth-screen">
      <section className="auth-card">
        <div className="brand">
          <div className="brand-mark">MDS</div>
          <div>
            <p className="eyebrow">Admin login</p>
            <h1>Message Delivery Service</h1>
          </div>
        </div>

        <p className="lede">
          Sign in with your admin username, password, and TOTP code. The first boot creates
          a default admin user and logs the credentials in the service output.
        </p>

        <form
          className="form-grid"
          onSubmit={async (event) => {
            event.preventDefault()
            setSubmitting(true)
            setMessage(null)
            try {
              await login({
                username: username.trim(),
                password,
                totpCode: totpCode.trim(),
              })
              navigate({ to: location.pathname })
            } catch (err) {
              setMessage(err instanceof Error ? err.message : 'Login failed')
            } finally {
              setSubmitting(false)
            }
          }}
        >
          <Field label="Username">
            <TextInput
              autoComplete="username"
              value={username}
              onChange={(event) => setUsername(event.target.value)}
            />
          </Field>
          <Field label="Password">
            <TextInput
              autoComplete="current-password"
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
            />
          </Field>
          <Field label="TOTP code" hint="6 digits from your authenticator app">
            <TextInput
              inputMode="numeric"
              autoComplete="one-time-code"
              value={totpCode}
              onChange={(event) => setTotpCode(event.target.value)}
            />
          </Field>
          {message || error ? <div className="feedback">{message ?? error}</div> : null}
          <div className="form-actions">
            <Button variant="primary" type="submit" disabled={submitting || status === 'loading'}>
              {submitting ? 'Signing in...' : 'Sign in'}
            </Button>
          </div>
        </form>
      </section>
    </div>
  )
}

export function UsersPage() {
  const { token } = useAdminAuth()
  const [users, setUsers] = useState<Awaited<ReturnType<typeof listAdminUsers>>['users']>([])
  const [username, setUsername] = useState('')
  const [status, setStatus] = useState<'idle' | 'loading' | 'saving'>('idle')
  const [error, setError] = useState<string | null>(null)
  const [credentials, setCredentials] = useState<{
    username: string
    password: string
    totpSecret: string
    provisioningUri: string
  } | null>(null)

  useEffect(() => {
    let cancelled = false
    async function loadUsers() {
      if (!token) {
        return
      }
      setStatus('loading')
      try {
        const response = await listAdminUsers(token)
        if (!cancelled) {
          setUsers(response.users)
          setError(null)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load users')
        }
      } finally {
        if (!cancelled) {
          setStatus('idle')
        }
      }
    }

    void loadUsers()

    return () => {
      cancelled = true
    }
  }, [token])

  if (!token) {
    return <LoginPage />
  }

  return (
    <div className="page">
      <header className="page-header">
        <div>
          <p className="eyebrow">Admin users</p>
          <h1>Username, password, and TOTP access</h1>
          <p className="lede">
            Create additional admin users for the console. Every new user gets a generated
            password and TOTP secret.
          </p>
        </div>
      </header>

      <section className="split-grid">
        <Panel title="Create user" description="Generate credentials for another admin account.">
          <form
            className="form-grid"
            onSubmit={async (event) => {
              event.preventDefault()
              if (!username.trim()) {
                return
              }
              setStatus('saving')
              setError(null)
              try {
                const response = await createAdminUser(token, username.trim())
                setUsers((current) => [...current, response.user].sort((a, b) => a.username.localeCompare(b.username)))
                setCredentials({
                  username: response.user.username,
                  password: response.credentials.password,
                  totpSecret: response.credentials.totpSecret,
                  provisioningUri: response.credentials.provisioningUri,
                })
                setUsername('')
              } catch (err) {
                setError(err instanceof Error ? err.message : 'Failed to create user')
              } finally {
                setStatus('idle')
              }
            }}
          >
            <Field label="Username" hint="Must be unique">
              <TextInput value={username} onChange={(event) => setUsername(event.target.value)} />
            </Field>
            <div className="form-actions">
              <Button variant="primary" type="submit" disabled={status === 'saving'}>
                {status === 'saving' ? 'Creating...' : 'Create user'}
              </Button>
            </div>
          </form>
        </Panel>

        <Panel title="Generated credentials" description="Shown once after user creation.">
          {credentials ? (
            <div className="stack">
              <div className="credential-block">
                <span>Username</span>
                <strong>{credentials.username}</strong>
              </div>
              <div className="credential-block">
                <span>Password</span>
                <strong className="inline-code">{credentials.password}</strong>
              </div>
              <div className="credential-block">
                <span>TOTP secret</span>
                <strong className="inline-code">{credentials.totpSecret}</strong>
              </div>
              <div className="credential-block">
                <span>Provisioning URI</span>
                <strong className="inline-code">{credentials.provisioningUri}</strong>
              </div>
            </div>
          ) : (
            <p className="subtle">Create a user to see the generated one-time credentials here.</p>
          )}
        </Panel>
      </section>

      <Panel
        title="Existing users"
        description={status === 'loading' ? 'Loading from the backend...' : 'Current admin accounts.'}
      >
        {error ? <div className="feedback">{error}</div> : null}
        <div className="table-wrap">
          <table className="table">
            <thead>
              <tr>
                <th>Username</th>
                <th>Created</th>
                <th>Last login</th>
              </tr>
            </thead>
            <tbody>
              {users.map((user) => (
                <tr key={user.id}>
                  <td>{user.username}</td>
                  <td>{formatDate(user.createdAt)}</td>
                  <td>{formatDate(user.lastLoginAt)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Panel>
    </div>
  )
}

function Panel({
  title,
  description,
  action,
  children,
}: {
  title: string
  description?: string
  action?: ReactNode
  children: ReactNode
}) {
  return (
    <section className="panel">
      <header className="panel-header">
        <div>
          <h3>{title}</h3>
          {description ? <p>{description}</p> : null}
        </div>
        {action}
      </header>
      {children}
    </section>
  )
}

function MetricCard({
  label,
  value,
  detail,
  tone = 'neutral',
}: {
  label: string
  value: string | number
  detail: string
  tone?: BadgeTone
}) {
  return (
    <article className="metric-card">
      <span className={toneClass(tone)}>{label}</span>
      <strong>{value}</strong>
      <p>{detail}</p>
    </article>
  )
}

function Field({
  label,
  hint,
  children,
}: {
  label: string
  hint?: string
  children: ReactNode
}) {
  return (
    <label className="field">
      <span className="field-label">
        {label}
        {hint ? <small>{hint}</small> : null}
      </span>
      {children}
    </label>
  )
}

function TextInput(
  props: InputHTMLAttributes<HTMLInputElement> & { variant?: 'default' | 'mono' },
) {
  const { variant = 'default', ...rest } = props

  return <input {...rest} className={`input ${variant === 'mono' ? 'mono' : ''}`} />
}

function SelectInput(props: SelectHTMLAttributes<HTMLSelectElement>) {
  return <select {...props} className="input" />
}

function TextArea(props: TextareaHTMLAttributes<HTMLTextAreaElement>) {
  return <textarea {...props} className="textarea" />
}

function Button({
  variant = 'secondary',
  children,
  className,
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: 'primary' | 'secondary' | 'danger'
}) {
  return (
    <button {...props} className={`button button-${variant} ${className ?? ''}`.trim()}>
      {children}
    </button>
  )
}

function Badge({
  tone = 'neutral',
  children,
}: {
  tone?: BadgeTone
  children: ReactNode
}) {
  return <span className={toneClass(tone)}>{children}</span>
}

function serviceSummary(scope: ServiceScope) {
  if (scope === 'all') {
    return 'Email + SMS'
  }

  return scope.toUpperCase()
}

export function OverviewPage() {
  const { state } = useAdminStore()

  const activeServices = state.services.filter((service) => service.status === 'active')
  const defaultEmail = getDefaultEmailAccount(state.emailAccounts)
  const connectedSms = state.smsCredentials.status === 'connected'
  const queuedMessages = state.messages.filter((message) => message.status === 'queued').length

  return (
    <div className="page">
      <header className="page-header">
        <div>
          <p className="eyebrow">Overview</p>
          <h1>Delivery control room</h1>
          <p className="lede">
            Keep signed service access, SMTP accounts, 46elks credentials, and outbound
            messages in one place.
          </p>
        </div>
        <div className="page-actions">
          <Link to="/send-message" className="button button-primary">
            Send a message
          </Link>
          <Link to="/services" className="button button-secondary">
            Review access
          </Link>
        </div>
      </header>

      <section className="metric-grid">
        <MetricCard
          label="Services"
          value={state.services.length}
          detail={`${activeServices.length} currently active and able to sign requests.`}
          tone="success"
        />
        <MetricCard
          label="SMTP accounts"
          value={state.emailAccounts.length}
          detail={
            defaultEmail
              ? `Default sender: ${defaultEmail.address}.`
              : 'No default sender configured yet.'
          }
          tone="info"
        />
        <MetricCard
          label="46elks"
          value={connectedSms ? 'Connected' : 'Stale'}
          detail={`Rotation count ${state.smsCredentials.rotationCount}.`}
          tone={connectedSms ? 'success' : 'warning'}
        />
        <MetricCard
          label="Queued messages"
          value={queuedMessages}
          detail="Queued messages from the backend."
          tone="neutral"
        />
      </section>

      <section className="dashboard-grid">
        <Panel
          title="Authorized services"
          description="Signing keys, scopes, and operational status."
          action={
            <Link className="text-link" to="/services">
              Open services
            </Link>
          }
        >
          <div className="stack">
            {state.services.slice(0, 3).map((service) => (
              <article key={service.id} className="mini-row">
                <div>
                  <strong>{service.name}</strong>
                  <p>{service.owner}</p>
                </div>
                <div className="mini-row-meta">
                  <Badge tone={service.status === 'active' ? 'success' : 'warning'}>
                    {service.status}
                  </Badge>
                  <Badge tone="neutral">{serviceSummary(service.scope)}</Badge>
                </div>
              </article>
            ))}
          </div>
        </Panel>

        <Panel
          title="Recent activity"
          description="Recent audit events from the backend."
          action={
            <Link className="text-link" to="/email-accounts">
              Inspect providers
            </Link>
          }
        >
          <div className="stack">
            {state.activity.slice(0, 4).map((entry) => (
              <article key={entry.id} className="activity-row">
                <div className={`activity-dot activity-dot-${entry.tone}`} />
                <div>
                  <strong>{entry.title}</strong>
                  <p>{entry.detail}</p>
                </div>
                <time>{formatDate(entry.createdAt)}</time>
              </article>
            ))}
          </div>
        </Panel>
      </section>

      <Panel
        title="Recent sends"
        description="The latest deliveries queued through the backend."
        action={
          <Link className="text-link" to="/send-message">
            Compose new message
          </Link>
        }
      >
        <div className="table-wrap">
          <table className="table">
            <thead>
              <tr>
                <th>Channel</th>
                <th>Service</th>
                <th>Recipients</th>
                <th>Sender</th>
                <th>Created</th>
              </tr>
            </thead>
            <tbody>
              {state.messages.slice(0, 4).map((message) => {
                const service = state.services.find((item) => item.id === message.serviceId)

                return (
                  <tr key={message.id}>
                    <td>
                      <Badge tone={message.channel === 'email' ? 'info' : 'warning'}>
                        {channelLabels[message.channel]}
                      </Badge>
                    </td>
                    <td>{service?.name ?? message.serviceId}</td>
                    <td>{message.recipients.join(', ')}</td>
                    <td>{message.sender}</td>
                    <td>{formatDate(message.createdAt)}</td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      </Panel>
    </div>
  )
}

export function ServicesPage() {
  const { state, dispatch } = useAdminStore()

  const [name, setName] = useState('')
  const [owner, setOwner] = useState('Platform Team')
  const [scope, setScope] = useState<ServiceScope>('all')
  const [status, setStatus] = useState<ServiceStatus>('active')
  const [notes, setNotes] = useState('')
  const [publicKey, setPublicKey] = useState('')

  return (
    <div className="page">
      <header className="page-header">
        <div>
          <p className="eyebrow">Services</p>
          <h1>Signed service access</h1>
          <p className="lede">
            Add, delete, reroll keys, and update channel scope for each authorized service.
          </p>
        </div>
      </header>

      <section className="split-grid">
        <Panel title="Add service" description="Create a new signing client and assign its scope.">
          <form
            className="form-grid"
            onSubmit={(event) => {
              event.preventDefault()

              const trimmedName = name.trim()
              const trimmedPublicKey = publicKey.trim()
              if (!trimmedName || !trimmedPublicKey) {
                return
              }

              const id = slugify(trimmedName) || `service-${state.services.length + 1}`

              dispatch({
                type: 'service/add',
                payload: {
                  id,
                  name: trimmedName,
                  owner: owner.trim() || 'Platform Team',
                  scope,
                  status,
                  publicKey: trimmedPublicKey,
                  notes: notes.trim(),
                },
              })

              setName('')
              setOwner('Platform Team')
              setScope('all')
              setStatus('active')
              setNotes('')
              setPublicKey('')
            }}
          >
            <Field label="Service name">
              <TextInput value={name} onChange={(event) => setName(event.target.value)} />
            </Field>
            <Field label="Owner / team">
              <TextInput value={owner} onChange={(event) => setOwner(event.target.value)} />
            </Field>
            <Field label="Channel scope">
              <SelectInput value={scope} onChange={(event) => setScope(event.target.value as ServiceScope)}>
                <option value="all">Email + SMS</option>
                <option value="email">Email only</option>
                <option value="sms">SMS only</option>
              </SelectInput>
            </Field>
            <Field label="Status">
              <SelectInput
                value={status}
                onChange={(event) => setStatus(event.target.value as ServiceStatus)}
              >
                <option value="active">Active</option>
                <option value="paused">Paused</option>
              </SelectInput>
            </Field>
            <Field label="Public key" hint="Ed25519 signing key">
              <TextArea
                rows={4}
                value={publicKey}
                onChange={(event) => setPublicKey(event.target.value)}
                placeholder="ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA..."
              />
            </Field>
            <Field label="Notes">
              <TextArea rows={3} value={notes} onChange={(event) => setNotes(event.target.value)} />
            </Field>
            <div className="form-actions">
              <Button variant="primary" type="submit">
                Add service
              </Button>
            </div>
          </form>
        </Panel>

        <Panel
          title="Registered services"
          description={`${state.services.length} service identities currently configured.`}
        >
          <div className="table-wrap">
            <table className="table table-compact">
              <thead>
                <tr>
                  <th>Service</th>
                  <th>Owner</th>
                  <th>Scope</th>
                  <th>Status</th>
                  <th>Key</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {state.services.map((service) => (
                  <tr key={service.id}>
                    <td>
                      <strong>{service.name}</strong>
                      <p className="subtle">{service.id}</p>
                    </td>
                    <td>{service.owner}</td>
                    <td>
                      <SelectInput
                        value={service.scope}
                        onChange={(event) =>
                          dispatch({
                            type: 'service/set-scope',
                            payload: { id: service.id, scope: event.target.value as ServiceScope },
                          })
                        }
                      >
                        <option value="all">Email + SMS</option>
                        <option value="email">Email only</option>
                        <option value="sms">SMS only</option>
                      </SelectInput>
                    </td>
                    <td>
                      <Button
                        type="button"
                        onClick={() =>
                          dispatch({
                            type: 'service/set-status',
                            payload: {
                              id: service.id,
                              status: service.status === 'active' ? 'paused' : 'active',
                            },
                          })
                        }
                      >
                        {service.status === 'active' ? 'Pause' : 'Resume'}
                      </Button>
                    </td>
                    <td>
                      <code className="inline-code">{service.publicKey.slice(0, 42)}…</code>
                    </td>
                    <td>
                      <div className="row-actions">
                        <Button
                          type="button"
                          onClick={() =>
                            dispatch({
                              type: 'service/reroll',
                              payload: { id: service.id },
                            })
                          }
                        >
                          Reroll key
                        </Button>
                        <Button
                          type="button"
                          variant="danger"
                          onClick={() => {
                            if (
                              window.confirm(`Delete service "${service.name}" from the registry?`)
                            ) {
                              dispatch({
                                type: 'service/delete',
                                payload: { id: service.id },
                              })
                            }
                          }}
                        >
                          Delete
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Panel>
      </section>
    </div>
  )
}

export function EmailAccountsPage() {
  const { state, dispatch } = useAdminStore()

  const [displayName, setDisplayName] = useState('')
  const [address, setAddress] = useState('')
  const [smtpHost, setSmtpHost] = useState('')
  const [smtpPort, setSmtpPort] = useState('587')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')

  return (
    <div className="page">
      <header className="page-header">
        <div>
          <p className="eyebrow">Email accounts</p>
          <h1>SMTP sender administration</h1>
          <p className="lede">
            Add senders, set the default account, and test connection state from one screen.
          </p>
        </div>
      </header>

      <section className="split-grid">
        <Panel title="Add SMTP account" description="Register another email sender profile.">
          <form
            className="form-grid"
            onSubmit={(event) => {
              event.preventDefault()

              const trimmedAddress = address.trim()
              const trimmedPassword = password.trim()
              if (!trimmedAddress || !trimmedPassword) {
                return
              }

              const id = slugify(trimmedAddress) || `email-${state.emailAccounts.length + 1}`

              dispatch({
                type: 'email/add',
                payload: {
                  id,
                  address: trimmedAddress,
                  displayName: displayName.trim() || trimmedAddress,
                  smtpHost: smtpHost.trim(),
                  smtpPort: Number(smtpPort) || 587,
                  username: username.trim() || trimmedAddress,
                  password: trimmedPassword,
                },
              })

              setDisplayName('')
              setAddress('')
              setSmtpHost('')
              setSmtpPort('587')
              setUsername('')
              setPassword('')
            }}
          >
            <Field label="Display name">
              <TextInput
                value={displayName}
                onChange={(event) => setDisplayName(event.target.value)}
              />
            </Field>
            <Field label="Email address">
              <TextInput value={address} onChange={(event) => setAddress(event.target.value)} />
            </Field>
            <Field label="SMTP host">
              <TextInput value={smtpHost} onChange={(event) => setSmtpHost(event.target.value)} />
            </Field>
            <Field label="SMTP port">
              <TextInput
                value={smtpPort}
                onChange={(event) => setSmtpPort(event.target.value)}
                inputMode="numeric"
              />
            </Field>
            <Field label="Username">
              <TextInput value={username} onChange={(event) => setUsername(event.target.value)} />
            </Field>
            <Field label="Password">
              <TextInput
                type="password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                placeholder="Enter SMTP password"
              />
            </Field>
            <div className="form-actions">
              <Button variant="primary" type="submit">
                Add account
              </Button>
            </div>
          </form>
        </Panel>

        <div className="stack">
          {state.emailAccounts.map((account) => (
            <Panel
              key={account.id}
              title={account.displayName}
              description={account.address}
              action={account.isDefault ? <Badge tone="success">Default</Badge> : null}
            >
              <div className="card-grid">
                <div>
                  <p className="field-label">SMTP endpoint</p>
                  <strong>
                    {account.smtpHost}:{account.smtpPort}
                  </strong>
                </div>
                <div>
                  <p className="field-label">Username</p>
                  <strong>{account.username}</strong>
                </div>
                <div>
                  <p className="field-label">Password</p>
                  <strong className="secret">{maskSecret(account.password)}</strong>
                </div>
                <div>
                  <p className="field-label">Status</p>
                  <Badge tone={account.status === 'healthy' ? 'success' : 'warning'}>
                    {account.status}
                  </Badge>
                </div>
                <div>
                  <p className="field-label">Last test</p>
                  <strong>{formatDate(account.lastTestedAt)}</strong>
                </div>
              </div>

              <div className="row-actions row-actions-wrap">
                <Button
                  type="button"
                  variant="secondary"
                  onClick={() =>
                    dispatch({
                      type: 'email/set-default',
                      payload: { id: account.id },
                    })
                  }
                >
                  Set default
                </Button>
                <Button
                  type="button"
                  onClick={() =>
                    dispatch({
                      type: 'email/test',
                      payload: { id: account.id },
                    })
                  }
                >
                  Test connection
                </Button>
                <Button
                  type="button"
                  variant="danger"
                  onClick={() => {
                    if (window.confirm(`Delete email account "${account.address}"?`)) {
                      dispatch({
                        type: 'email/delete',
                        payload: { id: account.id },
                      })
                    }
                  }}
                >
                  Delete
                </Button>
              </div>
            </Panel>
          ))}
        </div>
      </section>
    </div>
  )
}

export function FortySixElksPage() {
  const { state, dispatch } = useAdminStore()
  const [username, setUsername] = useState(state.smsCredentials.username)
  const [password, setPassword] = useState(state.smsCredentials.password)
  const [showPassword, setShowPassword] = useState(false)

  useEffect(() => {
    setUsername(state.smsCredentials.username)
    setPassword(state.smsCredentials.password)
  }, [state.smsCredentials.password, state.smsCredentials.username])

  return (
    <div className="page">
      <header className="page-header">
        <div>
          <p className="eyebrow">46elks</p>
          <h1>SMS provider credentials</h1>
          <p className="lede">
            Maintain the 46elks username and password used by the messaging backend.
          </p>
        </div>
      </header>

      <section className="split-grid">
        <Panel title="Credential editor" description="Update the 46elks credential pair.">
          <form
            className="form-grid"
            onSubmit={(event) => {
              event.preventDefault()

              const trimmedUsername = username.trim()
              const trimmedPassword = password.trim()
              if (!trimmedUsername || !trimmedPassword) {
                return
              }

              dispatch({
                type: 'sms/update',
                payload: {
                  username: trimmedUsername,
                  password: trimmedPassword,
                },
              })
            }}
          >
            <Field label="Username">
              <TextInput value={username} onChange={(event) => setUsername(event.target.value)} />
            </Field>
            <Field label="Password">
              <TextInput
                type={showPassword ? 'text' : 'password'}
                value={password}
                onChange={(event) => setPassword(event.target.value)}
              />
            </Field>
            <div className="form-actions">
              <Button type="button" onClick={() => setShowPassword((current) => !current)}>
                {showPassword ? 'Hide password' : 'Show password'}
              </Button>
              <Button variant="primary" type="submit">
                Save credentials
              </Button>
              <Button
                type="button"
                variant="secondary"
                onClick={() => dispatch({ type: 'sms/rotate' })}
              >
                Rotate secret
              </Button>
            </div>
          </form>
        </Panel>

        <div className="stack">
          <Panel title="Connection status" description="Current health for the SMS provider.">
            <div className="card-grid">
              <div>
                <p className="field-label">Provider</p>
                <strong>46elks</strong>
              </div>
              <div>
                <p className="field-label">Status</p>
                <Badge tone={state.smsCredentials.status === 'connected' ? 'success' : 'warning'}>
                  {state.smsCredentials.status}
                </Badge>
              </div>
              <div>
                <p className="field-label">Rotation count</p>
                <strong>{state.smsCredentials.rotationCount}</strong>
              </div>
              <div>
                <p className="field-label">Last synced</p>
                <strong>{formatDate(state.smsCredentials.lastSyncedAt)}</strong>
              </div>
            </div>
          </Panel>

          <Panel title="Recent SMS activity" description="A compact audit of provider changes.">
            <div className="stack">
              {state.activity
                .filter((entry) => entry.title.includes('46elks') || entry.title.includes('SMS'))
                .slice(0, 4)
                .map((entry) => (
                  <article key={entry.id} className="activity-row">
                    <div className={`activity-dot activity-dot-${entry.tone}`} />
                    <div>
                      <strong>{entry.title}</strong>
                      <p>{entry.detail}</p>
                    </div>
                    <time>{formatDate(entry.createdAt)}</time>
                  </article>
                ))}
            </div>
          </Panel>
        </div>
      </section>
    </div>
  )
}

function buildAdminMessageRequest({
  channel,
  serviceId,
  recipients,
  from,
  subject,
  senderName,
  contentMode,
  body,
  templateName,
  templateData,
}: {
  channel: 'email' | 'sms'
  serviceId: string
  recipients: string[]
  from: string
  subject: string
  senderName: string
  contentMode: 'plain' | 'html' | 'template'
  body: string
  templateName: string
  templateData: string
}): AdminMessageRequest {
  return {
    channel,
    serviceId,
    recipients,
    from: channel === 'email' ? from : undefined,
    senderName: channel === 'sms' ? senderName : undefined,
    subject: channel === 'email' ? subject : undefined,
    contentMode,
    body: contentMode === 'template' ? undefined : body,
    template:
      contentMode === 'template'
        ? {
            name: templateName,
            data: parseJsonOrRaw(templateData),
          }
        : undefined,
  }
}

function parseJsonOrRaw(value: string) {
  if (!value.trim()) {
    return {}
  }

  try {
    return JSON.parse(value) as Record<string, unknown>
  } catch {
    return { raw: value }
  }
}

export function SendMessagePage() {
  const { state, refresh } = useAdminStore()
  const { token } = useAdminAuth()
  const [channel, setChannel] = useState<'email' | 'sms'>('email')
  const [serviceId, setServiceId] = useState(state.services[0]?.id ?? '')
  const [recipients, setRecipients] = useState('')
  const [from, setFrom] = useState(getDefaultEmailAccount(state.emailAccounts)?.address ?? '')
  const [subject, setSubject] = useState('Welcome to the platform')
  const [senderName, setSenderName] = useState('MessageSvc')
  const [body, setBody] = useState('Hello from the message delivery service.')
  const [templateName, setTemplateName] = useState('welcome-message')
  const [templateData, setTemplateData] = useState('{\n  "name": "Alex"\n}')
  const [contentMode, setContentMode] = useState<'plain' | 'html' | 'template'>('plain')
  const [feedback, setFeedback] = useState<string | null>(null)
  const [preview, setPreview] = useState<{
    rendered: string
    warnings?: string[]
  } | null>(null)
  const [previewError, setPreviewError] = useState<string | null>(null)
  const [previewStatus, setPreviewStatus] = useState<'idle' | 'loading'>('idle')

  useEffect(() => {
    if (!state.services.some((service) => service.id === serviceId)) {
      setServiceId(state.services[0]?.id ?? '')
    }
  }, [serviceId, state.services])

  useEffect(() => {
    if (channel !== 'email') {
      return
    }

    const defaultSender = getDefaultEmailAccount(state.emailAccounts)?.address ?? ''
    if (!from && defaultSender) {
      setFrom(defaultSender)
    }
  }, [channel, from, state.emailAccounts])

  const request = buildAdminMessageRequest({
    channel,
    serviceId,
    recipients: splitValues(recipients),
    from,
    subject,
    senderName,
    contentMode,
    body,
    templateName,
    templateData,
  })

  const previewKey = JSON.stringify(request)

  useEffect(() => {
    let cancelled = false

    async function loadPreview() {
      if (!token || !serviceId || request.recipients.length === 0) {
        if (!cancelled) {
          setPreview(null)
          setPreviewError(null)
          setPreviewStatus('idle')
        }
        return
      }

      setPreviewStatus('loading')
      try {
        const response = await previewAdminMessage(token, request)
        if (!cancelled) {
          setPreview(response.preview)
          setPreviewError(null)
          setPreviewStatus('idle')
        }
      } catch (error) {
        if (!cancelled) {
          setPreview(null)
          setPreviewError(error instanceof Error ? error.message : 'Failed to preview message')
          setPreviewStatus('idle')
        }
      }
    }

    const handle = window.setTimeout(() => {
      void loadPreview()
    }, 250)

    return () => {
      cancelled = true
      window.clearTimeout(handle)
    }
  }, [previewKey, serviceId, token])

  return (
    <div className="page">
      <header className="page-header">
        <div>
          <p className="eyebrow">Send message</p>
          <h1>Compose delivery requests</h1>
          <p className="lede">
            Build the request payload in the UI, inspect the backend preview, and queue a send.
          </p>
        </div>
      </header>

      <section className="split-grid split-grid-wide">
        <Panel title="Composer" description="Switch between email and SMS without leaving the page.">
          <form
            className="form-grid"
            onSubmit={async (event) => {
              event.preventDefault()
              if (!token) {
                return
              }

              const recipientList = splitValues(recipients)
              if (recipientList.length === 0 || !serviceId) {
                return
              }

              await createAdminMessage(token, {
                ...request,
                recipients: recipientList,
                serviceId,
                body: request.body,
              })
              await refresh()

              setFeedback(
                `${channelLabels[channel]} message queued for ${recipientList.length} recipient(s).`,
              )
            }}
          >
            <Field label="Service identity">
              <SelectInput value={serviceId} onChange={(event) => setServiceId(event.target.value)}>
                {state.services.map((service) => (
                  <option key={service.id} value={service.id}>
                    {service.name} ({service.id})
                  </option>
                ))}
              </SelectInput>
            </Field>

            <div className="segmented">
              <button
                type="button"
                className={channel === 'email' ? 'segmented-item is-active' : 'segmented-item'}
                onClick={() => setChannel('email')}
              >
                Email
              </button>
              <button
                type="button"
                className={channel === 'sms' ? 'segmented-item is-active' : 'segmented-item'}
                onClick={() => setChannel('sms')}
              >
                SMS
              </button>
            </div>

            <Field
              label={channel === 'email' ? 'Recipient email addresses' : 'Recipient phone numbers'}
              hint="Comma or newline separated"
            >
              <TextArea
                rows={3}
                value={recipients}
                onChange={(event) => setRecipients(event.target.value)}
                placeholder={channel === 'email' ? 'one@example.com, two@example.com' : '+46700000000, +46700000001'}
              />
            </Field>

            {channel === 'email' ? (
              <>
                <Field label="From address">
                  <SelectInput value={from} onChange={(event) => setFrom(event.target.value)}>
                    {state.emailAccounts.map((account) => (
                      <option key={account.id} value={account.address}>
                        {account.displayName} ({account.address})
                      </option>
                    ))}
                  </SelectInput>
                </Field>
                <Field label="Subject">
                  <TextInput value={subject} onChange={(event) => setSubject(event.target.value)} />
                </Field>
              </>
            ) : (
              <Field label="Sender name" hint="Must fit SMS provider limits">
                <TextInput
                  value={senderName}
                  onChange={(event) => setSenderName(event.target.value)}
                  maxLength={11}
                />
              </Field>
            )}

            <Field label="Content mode">
              <SelectInput
                value={contentMode}
                onChange={(event) => setContentMode(event.target.value as 'plain' | 'html' | 'template')}
              >
                <option value="plain">Plain text</option>
                {channel === 'email' ? <option value="html">HTML</option> : null}
                <option value="template">Template</option>
              </SelectInput>
            </Field>

            {contentMode === 'template' ? (
              <>
                <Field label="Template name">
                  <TextInput
                    value={templateName}
                    onChange={(event) => setTemplateName(event.target.value)}
                  />
                </Field>
                <Field label="Template data" hint="JSON payload">
                  <TextArea
                    rows={5}
                    value={templateData}
                    onChange={(event) => setTemplateData(event.target.value)}
                  />
                </Field>
              </>
            ) : (
              <Field label="Message body">
                <TextArea rows={6} value={body} onChange={(event) => setBody(event.target.value)} />
              </Field>
            )}

            {feedback ? <div className="feedback">{feedback}</div> : null}

            <div className="form-actions">
              <Button variant="primary" type="submit">
                Queue message
              </Button>
            </div>
          </form>
        </Panel>

        <div className="stack">
          <Panel
            title="Payload preview"
            description={previewStatus === 'loading' ? 'Fetching backend preview...' : 'Backend-rendered payload preview.'}
          >
            {previewError ? <div className="feedback">{previewError}</div> : null}
            {preview ? (
              <pre className="preview-code">{JSON.stringify({ request, preview }, null, 2)}</pre>
            ) : (
              <pre className="preview-code">{JSON.stringify(request, null, 2)}</pre>
            )}
          </Panel>

          <Panel title="Recent messages" description="Recent deliveries from the backend.">
            <div className="stack">
              {state.messages.slice(0, 6).map((message) => (
                <article key={message.id} className="message-row">
                  <div className="message-row-head">
                    <div>
                      <strong>{message.subject ?? message.templateName ?? 'Untitled message'}</strong>
                      <p>{message.recipients.join(', ')}</p>
                    </div>
                    <Badge tone={message.channel === 'email' ? 'info' : 'warning'}>
                      {channelLabels[message.channel]}
                    </Badge>
                  </div>
                  <div className="message-row-meta">
                    <span>{message.sender}</span>
                    <span>{formatDate(message.createdAt)}</span>
                  </div>
                </article>
              ))}
            </div>
          </Panel>
        </div>
      </section>
    </div>
  )
}
