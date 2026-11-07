import { useEffect, useState } from 'react'
import {
  Activity,
  ArrowRight,
  ArrowUpRight,
  Box,
  FolderKanban,
  LayoutDashboard,
  LogOut,
  Rocket,
  Server,
  ShieldCheck,
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { authenticateSuperuser, getOnboardingStatus, refreshSuperuser } from '@/lib/pocketbase'

const destinations = [
  { id: 'overview', label: 'Overview', icon: LayoutDashboard, index: '01' },
  { id: 'servers', label: 'Servers', icon: Server, index: '02' },
  { id: 'projects', label: 'Projects', icon: FolderKanban, index: '03' },
  { id: 'deployments', label: 'Deployments', icon: Rocket, index: '04' },
]

const routeDetails = {
  servers: {
    number: '02',
    eyebrow: 'Infrastructure / remote hosts',
    title: 'Servers',
    description: 'Linux hosts connected over SSH will be tracked here.',
    columns: ['SERVER', 'ADDRESS', 'DOCKER', 'READINESS'],
    emptyTitle: 'No server records',
    emptyCopy: 'SSH registration and Docker checks are the next infrastructure workflow to land.',
    icon: Server,
  },
  projects: {
    number: '03',
    eyebrow: 'Workspace / application groups',
    title: 'Projects',
    description: 'Keep related applications together in a single workspace.',
    columns: ['PROJECT', 'SLUG', 'APPLICATIONS', 'UPDATED'],
    emptyTitle: 'No projects created',
    emptyCopy: 'Project records will organize applications after server management is available.',
    icon: FolderKanban,
  },
  deployments: {
    number: '04',
    eyebrow: 'Release log / compose operations',
    title: 'Deployments',
    description: 'Follow release state, output, and running containers from one place.',
    columns: ['RELEASE', 'APPLICATION', 'STATUS', 'STARTED'],
    emptyTitle: 'No deployment records',
    emptyCopy: 'Compose releases and their runtime history will show up here once the deployment worker is ready.',
    icon: Rocket,
  },
}

function currentRoute() {
  const route = window.location.hash.replace(/^#\/?/, '')
  return destinations.some(({ id }) => id === route) ? route : 'overview'
}

function Navigation({ active, mobile = false }) {
  return (
    <nav className={mobile ? 'mobile-navigation' : 'desktop-navigation'} aria-label="Primary navigation">
      {destinations.map(({ id, label, icon: Icon, index }) => (
        <a
          key={id}
          className={`nav-link${active === id ? ' is-active' : ''}`}
          href={`#/${id}`}
          aria-current={active === id ? 'page' : undefined}
        >
          <span className="nav-index" aria-hidden="true">{index}</span>
          <Icon className="nav-icon" aria-hidden="true" />
          <span className="nav-label">{label}</span>
          {active === id && <ArrowUpRight className="nav-current-mark" aria-hidden="true" />}
        </a>
      ))}
    </nav>
  )
}

function Header({ title, auth, onLogout }) {
  return (
    <header className="topbar">
      <div className="breadcrumbs" aria-label="Location">
        <span>FIELD OPS</span>
        <span className="breadcrumb-slash" aria-hidden="true">/</span>
        <span>{title.toUpperCase()}</span>
      </div>
      <div className="topbar-tools">
        <Badge className="preview-badge"><span aria-hidden="true">●</span> SUPERUSER SESSION</Badge>
        <span className="session-identity">{auth.record?.email || 'PocketBase superuser'}</span>
        <Button className="logout-button" variant="outline" onClick={onLogout}>
          <LogOut aria-hidden="true" /> Sign out
        </Button>
      </div>
    </header>
  )
}

function LoginScreen({ notice, onAuthenticated }) {
  const [identity, setIdentity] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [onboarding, setOnboarding] = useState(null)

  useEffect(() => {
    getOnboardingStatus().then(setOnboarding).catch(() => setOnboarding(null))
  }, [])

  async function handleSubmit(event) {
    event.preventDefault()
    if (submitting) return
    setError('')
    setSubmitting(true)
    try {
      const result = await authenticateSuperuser(identity.trim(), password)
      onAuthenticated(result)
    } catch (requestError) {
      if (requestError.status === 400 || requestError.status === 401) {
        setError('Email atau kata sandi tidak valid.')
      } else {
        setError('PocketBase tidak dapat dijangkau. Periksa koneksi lalu coba lagi.')
      }
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <main className="login-screen">
      <div className="login-panel">
        <a className="brand login-brand" href="#/overview" aria-label="A-R-S-Y">
          <span className="brand-burst" aria-hidden="true"><span>A</span></span>
          <span className="brand-word">A-R-S-Y<span className="brand-period">.</span></span>
        </a>
        <p className="section-index">CONTROL ROOM / ADMIN ACCESS</p>
        <h1>Sign in<span className="heading-period">.</span></h1>
        <p className="login-copy">Masuk dengan akun superuser PocketBase untuk membuka console A-R-S-Y.</p>
        {onboarding && !onboarding.superuser_exists && (
          <div className="setup-command">
            <strong>First install?</strong>
            <span>Buat superuser dari terminal host, lalu kembali ke sini.</span>
            <code>arsy superuser create EMAIL PASSWORD</code>
          </div>
        )}
        {(notice || error) && <p className="login-message" role="alert">{error || notice}</p>}
        <form className="login-form" onSubmit={handleSubmit}>
          <label htmlFor="superuser-identity">Email</label>
          <input
            autoComplete="username"
            className="login-input"
            id="superuser-identity"
            name="identity"
            onChange={(event) => setIdentity(event.target.value)}
            required
            type="email"
            value={identity}
          />
          <label htmlFor="superuser-password">Password</label>
          <input
            autoComplete="current-password"
            className="login-input"
            id="superuser-password"
            name="password"
            onChange={(event) => setPassword(event.target.value)}
            required
            type="password"
            value={password}
          />
          <Button className="riot-button login-submit" disabled={submitting} type="submit">
            {submitting ? 'Checking credentials…' : 'Enter the console'}
            {!submitting && <ArrowRight aria-hidden="true" />}
          </Button>
        </form>
        <p className="login-footnote">SESSION TOKEN STAYS IN MEMORY · AUTHENTICATED BY POCKETBASE</p>
      </div>
      <div className="login-stamp" aria-hidden="true">NO<br />GUESTS</div>
    </main>
  )
}

function ComicArtwork() {
  return (
    <div className="art-stage" aria-hidden="true">
      <div className="art-halftone" />
      <svg className="deploy-art" viewBox="0 0 440 330" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path className="burst-shape" d="M218 5L247 47L294 23L299 74L350 68L336 116L386 133L355 173L397 210L350 234L369 282L318 280L303 327L260 299L226 339L199 296L151 324L143 274L92 282L106 231L57 211L88 171L47 134L96 110L78 62L129 66L143 16L186 46L218 5Z" />
        <path d="M215 31L233 57L261 43L263 74L294 70L285 99L315 109L297 133L322 155L294 170L306 199L275 200L267 231L241 214L220 241L203 214L174 230L171 199L140 201L150 171L120 159L139 135L114 112L144 97L133 68L164 71L173 41L199 59L215 31Z" fill="#D5EB42" stroke="#171918" strokeWidth="5" />
        <g className="terminal-tilt" transform="rotate(-6 220 164)">
          <rect x="102" y="85" width="238" height="170" rx="4" fill="#F7F0DF" stroke="#171918" strokeWidth="7" />
          <path d="M105 90H337V122H105V90Z" fill="#F04B32" stroke="#171918" strokeWidth="5" />
          <circle cx="124" cy="106" r="4" fill="#171918" />
          <circle cx="140" cy="106" r="4" fill="#171918" />
          <circle cx="156" cy="106" r="4" fill="#171918" />
          <path d="M129 155L148 170L129 185" stroke="#171918" strokeWidth="9" strokeLinecap="square" strokeLinejoin="miter" />
          <path d="M162 185H221" stroke="#171918" strokeWidth="8" strokeLinecap="square" />
          <path d="M128 210H195" stroke="#387FC8" strokeWidth="7" strokeLinecap="square" />
          <path d="M210 210H286" stroke="#F04B32" strokeWidth="7" strokeLinecap="square" />
          <rect x="267" y="145" width="44" height="40" fill="#387FC8" stroke="#171918" strokeWidth="5" />
          <path d="M267 145L289 132L311 145M289 132V173" stroke="#171918" strokeWidth="5" strokeLinejoin="round" />
        </g>
        <path d="M334 69L354 43" stroke="#171918" strokeWidth="8" strokeLinecap="round" />
        <path d="M369 110L405 103" stroke="#171918" strokeWidth="8" strokeLinecap="round" />
        <path d="M75 245L45 260" stroke="#171918" strokeWidth="8" strokeLinecap="round" />
        <path d="M102 64L84 43" stroke="#171918" strokeWidth="8" strokeLinecap="round" />
        <path d="M341 258L364 276" stroke="#171918" strokeWidth="8" strokeLinecap="round" />
        <path d="M167 44L160 19" stroke="#171918" strokeWidth="8" strokeLinecap="round" />
        <g transform="rotate(8 335 226)">
          <rect x="320" y="209" width="79" height="37" rx="2" fill="#F04B32" stroke="#171918" strokeWidth="4" />
          <text x="359" y="233" textAnchor="middle" fill="#171918" fontFamily="Bangers, sans-serif" fontSize="21" letterSpacing="1">SHIP IT</text>
        </g>
      </svg>
      <span className="art-caption">COMPOSE / 01</span>
    </div>
  )
}

function OnboardingWizard({ status, error }) {
  if (error) {
    return <Card className="setup-card setup-error"><CardContent>Setup status could not be loaded. Refresh this page to retry.</CardContent></Card>
  }
  if (!status) {
    return <Card className="setup-card"><CardContent>Loading first-run checklist…</CardContent></Card>
  }

  const steps = [
    { id: '01', title: 'Superuser access', detail: 'PocketBase administrator authentication', done: status.superuser_exists },
    { id: '02', title: 'Encryption key', detail: 'Secret source is configured for stored credentials', done: status.secret_key_configured },
    { id: '03', title: 'First server', detail: 'Register a Linux host for remote operations', done: status.server_registered },
  ]
  const completed = steps.filter(({ done }) => done).length

  return (
    <Card className="setup-card" aria-labelledby="setup-title">
      <CardHeader className="setup-header">
        <div>
          <span className="section-index">FIRST RUN / CONTROL PLANE SETUP</span>
          <CardTitle id="setup-title">{completed === steps.length ? 'Ready for operations' : 'Bring the control room online'}</CardTitle>
        </div>
        <Badge className={completed === steps.length ? 'setup-progress is-complete' : 'setup-progress'}>{completed} / {steps.length} READY</Badge>
      </CardHeader>
      <CardContent className="setup-content">
        <p className="setup-intro">A-R-S-Y runs the control plane locally. Your application containers stay on the Linux servers you connect.</p>
        <ol className="setup-steps">
          {steps.map(({ id, title, detail, done }) => (
            <li className={done ? 'setup-step is-done' : 'setup-step'} key={id}>
              <span className="setup-step-id">{done ? '✓' : id}</span>
              <span className="setup-step-copy"><strong>{title}</strong><small>{detail}</small></span>
              <Badge className={done ? 'setup-state is-complete' : 'setup-state'}>{done ? 'READY' : 'NEXT'}</Badge>
            </li>
          ))}
        </ol>
        {!status.secret_key_configured && <p className="setup-hint">Set the configured secret-key environment variable before storing SSH credentials.</p>}
        {!status.server_registered && (
          <Button asChild className="riot-button setup-action">
            <a href="#/servers">Open server setup <ArrowRight aria-hidden="true" /></a>
          </Button>
        )}
      </CardContent>
    </Card>
  )
}

function Overview({ onboarding, onboardingError }) {
  return (
    <>
      <OnboardingWizard error={onboardingError} status={onboarding} />
      <section className="comic-hero" aria-labelledby="hero-title">
        <div className="hero-copy">
          <div className="hero-kicker-row">
            <span className="issue-tag">ISSUE 01</span>
            <span className="hero-kicker">SELF-HOSTED DEPLOYMENTS</span>
          </div>
          <h2 id="hero-title">OWN YOUR<br /><span>DEPLOYMENT.</span></h2>
          <p>A-R-S-Y is building a Go-native control room for remote servers and Docker Compose apps.</p>
          <Button asChild className="riot-button" size="lg">
            <a href="#/servers">Explore server setup <ArrowRight aria-hidden="true" /></a>
          </Button>
          <span className="hero-footnote">CONTROL PLANE / MVP-1 IN PROGRESS</span>
        </div>
        <ComicArtwork />
      </section>

      <section className="metrics-grid" aria-label="Current workspace records">
        <MetricCard number="0" label="SERVERS" detail="No remote hosts" tone="red" icon={Server} />
        <MetricCard number="0" label="PROJECTS" detail="No workspaces" tone="blue" icon={FolderKanban} />
        <MetricCard number="0" label="DEPLOYMENTS" detail="No releases yet" tone="lime" icon={Rocket} />
      </section>

      <div className="operations-grid">
        <Card className="operations-card">
          <CardHeader className="section-card-header">
            <div>
              <span className="section-index">A / BUILD BOARD</span>
              <CardTitle>Next up in the stack</CardTitle>
            </div>
            <Badge className="status-badge status-foundation">FOUNDATION</Badge>
          </CardHeader>
          <CardContent className="lane-list">
            <RoadmapRow index="01" title="Connect a Linux server" detail="Pinned SSH host key + Docker checks" status="NEXT" icon={ShieldCheck} tone="red" />
            <Separator className="ink-separator" />
            <RoadmapRow index="02" title="Group apps into projects" detail="One application, one target server" status="PLANNED" icon={FolderKanban} tone="blue" />
            <Separator className="ink-separator" />
            <RoadmapRow index="03" title="Ship a Compose release" detail="Existing images, tracked history" status="PLANNED" icon={Box} tone="lime" />
          </CardContent>
        </Card>

        <Card className="activity-card">
          <CardHeader className="section-card-header">
            <div>
              <span className="section-index">B / EVENT LOG</span>
              <CardTitle>Recent activity</CardTitle>
            </div>
            <Activity className="section-icon" aria-hidden="true" />
          </CardHeader>
          <CardContent className="activity-empty">
            <span className="activity-stamp" aria-hidden="true">LOG<br />000</span>
            <div>
              <strong>Quiet on the wire.</strong>
              <p>Deployment events will show here once the release workflow is in place.</p>
            </div>
          </CardContent>
        </Card>
      </div>
    </>
  )
}

function MetricCard({ number, label, detail, tone, icon: Icon }) {
  return (
    <Card className={`metric-card metric-${tone}`}>
      <CardContent className="metric-content">
        <div className="metric-topline">
          <span className="metric-label">{label}</span>
          <Icon className="metric-icon" aria-hidden="true" />
        </div>
        <div className="metric-bottomline">
          <span className="metric-number">{number}</span>
          <span className="metric-detail">{detail}</span>
        </div>
      </CardContent>
    </Card>
  )
}

function RoadmapRow({ index, title, detail, status, icon: Icon, tone }) {
  return (
    <div className="roadmap-row">
      <span className={`roadmap-icon icon-${tone}`}><Icon aria-hidden="true" /></span>
      <span className="roadmap-index">{index}</span>
      <div className="roadmap-copy">
        <strong>{title}</strong>
        <span>{detail}</span>
      </div>
      <Badge className={`roadmap-status status-${tone}`}>{status}</Badge>
    </div>
  )
}

function EmptyRoute({ detail }) {
  const Icon = detail.icon
  return (
    <>
      <div className="route-intro">
        <span className="issue-tag">FILE {detail.number}</span>
        <div>
          <p className="section-index">{detail.eyebrow.toUpperCase()}</p>
          <h2>{detail.title}</h2>
          <p className="route-description">{detail.description}</p>
        </div>
        <Icon className="route-icon" aria-hidden="true" />
      </div>
      <Card className="records-card">
        <table className="records-table" aria-label={`${detail.title} records`}>
          <thead>
            <tr>{detail.columns.map((column) => <th scope="col" key={column}>{column}</th>)}</tr>
          </thead>
          <tbody>
            <tr>
              <td colSpan={detail.columns.length}>
                <div className="records-empty">
                  <div className="empty-art" aria-hidden="true"><span>!</span><i /><i /><i /></div>
                  <h3>{detail.emptyTitle}</h3>
                  <p>{detail.emptyCopy}</p>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </Card>
      <p className="route-footnote"><span aria-hidden="true">*</span> THIS AREA WILL FILL WITH LIVE RECORDS AS MVP-1 FEATURES LAND.</p>
    </>
  )
}

export default function App() {
  const [active, setActive] = useState(currentRoute)
  const [auth, setAuth] = useState(null)
  const [sessionNotice, setSessionNotice] = useState('')
  const [onboarding, setOnboarding] = useState(null)
  const [onboardingError, setOnboardingError] = useState(false)

  useEffect(() => {
    const handleHashChange = () => {
      setActive(currentRoute())
      window.requestAnimationFrame(() => {
        document.querySelector('#page-title')?.focus({ preventScroll: true })
      })
    }
    window.addEventListener('hashchange', handleHashChange)
    return () => window.removeEventListener('hashchange', handleHashChange)
  }, [])

  useEffect(() => {
    if (!auth?.token) return undefined
    const refreshTimer = window.setInterval(async () => {
      try {
        const refreshed = await refreshSuperuser(auth.token)
        setAuth(refreshed)
      } catch (error) {
        if (error.status === 401) {
          setAuth(null)
          setSessionNotice('Sesi berakhir. Masuk kembali untuk melanjutkan.')
          window.location.hash = '#/login'
        }
      }
    }, 10 * 60 * 1000)
    return () => window.clearInterval(refreshTimer)
  }, [auth?.token])

  useEffect(() => {
    if (!auth?.token) return undefined
    let activeRequest = true
    getOnboardingStatus()
      .then((status) => {
        if (activeRequest) {
          setOnboarding(status)
          setOnboardingError(false)
        }
      })
      .catch(() => {
        if (activeRequest) setOnboardingError(true)
      })
    return () => { activeRequest = false }
  }, [auth?.token])

  const current = destinations.find(({ id }) => id === active) ?? destinations[0]
  const routeDetail = routeDetails[active]

  useEffect(() => {
    document.title = `${current.label} · A-R-S-Y Console`
  }, [current.label])

  if (!auth) {
    return (
      <LoginScreen
        notice={sessionNotice}
        onAuthenticated={(session) => {
          setSessionNotice('')
          setAuth(session)
          window.location.hash = '#/overview'
        }}
      />
    )
  }

  return (
    <>
      <a className="skip-link" href="#page-title">Skip to content</a>
      <div className="app-shell">
        <aside className="sidebar">
          <a className="brand" href="#/overview" aria-label="A-R-S-Y overview">
            <span className="brand-burst" aria-hidden="true"><span>A</span></span>
            <span className="brand-word">A-R-S-Y<span className="brand-period">.</span></span>
          </a>
          <div className="edition-strip"><span>CONTROL PLANE</span><span>VOL. 01</span></div>
          <p className="sidebar-section-label">OPERATIONS</p>
          <Navigation active={active} />
          <div className="sidebar-bottom">
            <span className="sidebar-rule" aria-hidden="true" />
            <p>RUN IT YOUR WAY.<br /><span>SELF-HOSTED / GO NATIVE</span></p>
            <span className="sidebar-coordinates">ARSY / 26.01</span>
          </div>
        </aside>

        <div className="workspace">
          <Header
            auth={auth}
            onLogout={() => {
              setAuth(null)
              setSessionNotice('Sesi telah berakhir. Masuk kembali untuk melanjutkan.')
              window.location.hash = '#/login'
            }}
            title={current.label}
          />
          <main className="main-content">
            <div className="page-heading">
              <div className="page-heading-copy">
                <p className="section-index">A-R-S-Y / COMMAND ROOM</p>
                <h1 id="page-title" tabIndex="-1">{current.label}<span className="heading-period">.</span></h1>
              </div>
              <span className="page-number">{current.index}<span> / 04</span></span>
            </div>
            {active === 'overview' ? <Overview onboarding={onboarding} onboardingError={onboardingError} /> : <EmptyRoute detail={routeDetail} />}
            <footer className="page-footer">
              <span>A-R-S-Y / INDEPENDENTLY MAINTAINED</span>
              <span>BUILT FOR THE LONG HAUL <span aria-hidden="true">✳</span></span>
            </footer>
          </main>
        </div>
        <Navigation active={active} mobile />
      </div>
    </>
  )
}
