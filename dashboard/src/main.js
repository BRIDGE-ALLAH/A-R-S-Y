import "./style.css";
import { currentUser, isAuthenticated, login, logout, verifySession } from "./auth.js";

const pages = {
  overview: ["Overview", "Your infrastructure at a glance."],
  servers: ["Servers", "Manage the Linux hosts connected to this control plane."],
  projects: ["Projects", "Group related applications into projects."],
  deployments: ["Deployments", "Review deployment activity and application status."],
};

let onboarding = null;
let loginError = "";
let sessionCheck = false;

async function loadOnboarding() {
  try {
    const response = await fetch("/api/arsy/v1/onboarding");
    if (!response.ok) throw new Error("The onboarding service is not available yet.");
    onboarding = await response.json();
  } catch (error) {
    onboarding = { unavailable: true, message: error.message };
  }
  render();
}

function renderFirstRun() {
  document.querySelector("#app").innerHTML = `
    <main class="auth-screen">
      <section class="auth-card">
        <div class="brand"><span class="brand-mark">A</span><span>A-R-S-Y</span></div>
        <div class="eyebrow">FIRST-RUN SETUP</div>
        <h1>Welcome to A-R-S-Y</h1>
        <p class="subtitle">A-R-S-Y is a self-hosted control plane. It runs on your control server and connects to managed Linux hosts over SSH.</p>
        <div class="step-card">
          <span class="step-number">01</span>
          <div><strong>Create the first superuser</strong><p>Use the server CLI. A-R-S-Y never exposes an unauthenticated public account-creation endpoint.</p><code>arsy superuser create EMAIL PASSWORD</code></div>
        </div>
        <p class="hint">After creating the superuser, refresh this page to continue with sign-in and server setup.</p>
      </section>
    </main>`;
}

function renderLogin() {
  document.querySelector("#app").innerHTML = `
    <main class="auth-screen">
      <section class="auth-card">
        <div class="brand"><span class="brand-mark">A</span><span>A-R-S-Y</span></div>
        <div class="eyebrow">CONTROL PLANE</div>
        <h1>Sign in</h1>
        <p class="subtitle">Use your existing PocketBase superuser account.</p>
        <form id="login-form" class="login-form">
          <label for="identity">Email</label>
          <input id="identity" name="identity" type="email" autocomplete="username" required />
          <label for="password">Password</label>
          <input id="password" name="password" type="password" autocomplete="current-password" required />
          <p id="login-error" class="form-error" role="alert"></p>
          <button type="submit" class="primary-button">Sign in</button>
        </form>
      </section>
    </main>`;
  const form = document.querySelector("#login-form");
  const error = document.querySelector("#login-error");
  error.textContent = loginError;
  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    const submit = form.querySelector("button");
    submit.disabled = true;
    submit.textContent = "Signing in…";
    error.textContent = "";
    const values = new FormData(form);
    try {
      await login(String(values.get("identity") || ""), String(values.get("password") || ""));
      loginError = "";
      sessionCheck = false;
      await render();
    } catch (err) {
      loginError = err.message || "Unable to sign in.";
      error.textContent = loginError;
      submit.disabled = false;
      submit.textContent = "Sign in";
    }
  });
}

function renderConsole() {
  const nav = Object.entries(pages)
    .map(([id, page]) => `<a class="nav-link" href="#${id}" data-page="${id}">${page[0]}</a>`)
    .join("");
  const user = currentUser();
  document.querySelector("#app").innerHTML = `
    <div class="shell">
      <aside class="sidebar">
        <a class="brand" href="#overview"><span class="brand-mark">A</span><span>A-R-S-Y</span></a>
        <p class="nav-label">WORKSPACE</p>
        <nav aria-label="Main navigation">${nav}</nav>
        <div class="sidebar-foot"><span class="status-dot"></span> Control plane</div>
      </aside>
      <main class="main">
        <header class="topbar">
          <span>Self-hosted deployment control plane</span>
          <div class="topbar-actions"><span class="user-label">${user?.email || "Superuser"}</span><button id="logout" class="quiet-button">Sign out</button></div>
        </header>
        <section id="page" class="page" aria-live="polite"></section>
      </main>
    </div>`;
  document.querySelector("#logout").addEventListener("click", () => {
    logout();
    loginError = "";
    render();
  });
  document.querySelectorAll(".nav-link").forEach((link) => {
    link.addEventListener("click", () => {
      setTimeout(renderPage, 0);
    });
  });
  renderPage();
}

function renderPage() {
  const id = location.hash.slice(1) || "overview";
  const page = pages[id] || pages.overview;
  document.querySelectorAll(".nav-link").forEach((link) => {
    link.classList.toggle("active", link.dataset.page === id);
    if (link.dataset.page === id) link.setAttribute("aria-current", "page");
    else link.removeAttribute("aria-current");
  });

  let content = "";
  if (id === "overview") {
    content = `
      <article class="status-grid">
        <div class="status-card"><span>Superuser</span><strong>${onboarding.superuser_configured ? "Configured" : "Required"}</strong></div>
        <div class="status-card"><span>Secret key</span><strong>${onboarding.secret_key_configured ? "Configured" : "Action required"}</strong></div>
        <div class="status-card"><span>Managed servers</span><strong>${onboarding.server_registered ? "Registered" : "Not registered"}</strong></div>
      </article>
      ${!onboarding.secret_key_configured ? '<article class="notice-card"><h2>Configure the secret key</h2><p>Set <code>ARSY_SECRET_KEY</code> in the control-plane process environment before storing SSH credentials. Keep the value outside source control.</p></article>' : ""}
      ${!onboarding.server_registered ? '<article class="empty-card"><h2>Register your first server</h2><p>Open Servers to continue setup. A-R-S-Y will require an SSH host-key fingerprint confirmation before trusting a host.</p></article>' : ""}
    `;
  } else if (id === "servers") {
    content = onboarding.server_registered
      ? '<article class="empty-card"><h2>Server records detected</h2><p>Server inspection and management actions are being added in the SSH management phase.</p></article>'
      : '<article class="empty-card"><h2>No servers registered</h2><p>First register a Linux host from this section. SSH host-key pinning is required before a connection is trusted.</p></article>';
  } else if (id === "projects") {
    content = '<article class="empty-card"><h2>No project workflow yet</h2><p>Projects will organize applications and their deployment history.</p></article>';
  } else {
    content = '<article class="empty-card"><h2>No deployments yet</h2><p>Deployment history will appear after an application is deployed with Docker Compose.</p></article>';
  }

  document.querySelector("#page").innerHTML = `
    <div class="eyebrow">A-R-S-Y CONSOLE</div>
    <h1>${page[0]}</h1>
    <p class="subtitle">${page[1]}</p>
    ${content}
  `;
}

async function render() {
  if (!onboarding) {
    document.querySelector("#app").innerHTML = '<main class="auth-screen"><section class="auth-card"><h1>Loading A-R-S-Y…</h1></section></main>';
    return;
  }
  if (onboarding.unavailable) {
    document.querySelector("#app").innerHTML = `<main class="auth-screen"><section class="auth-card"><h1>A-R-S-Y is starting</h1><p class="subtitle">${onboarding.message}</p><button id="retry" class="primary-button">Retry</button></section></main>`;
    document.querySelector("#retry").addEventListener("click", loadOnboarding);
    return;
  }
  if (!onboarding.superuser_configured) {
    renderFirstRun();
    return;
  }
  if (!isAuthenticated()) {
    renderLogin();
    return;
  }
  if (!sessionCheck) {
    sessionCheck = true;
    const valid = await verifySession();
    if (!valid) {
      loginError = "Your session expired. Please sign in again.";
      renderLogin();
      return;
    }
  }
  renderConsole();
}

window.addEventListener("hashchange", renderPage);
loadOnboarding();
