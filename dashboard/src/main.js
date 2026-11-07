import "./style.css";
import { apiFetch, currentUser, isAuthenticated, login, logout, verifySession } from "./auth.js";

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
    content = '<div id="server-manager"><article class="empty-card"><h2>Loading servers…</h2></article></div>';
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
  if (id === "servers") void renderServerManager();
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

function renderServerManager() {
  const root = document.querySelector("#server-manager");
  if (!root) return;
  root.innerHTML = `
    <section class="server-form-card">
      <h2>Register a Linux server</h2>
      <p>Use an unencrypted OpenSSH private key. The key is encrypted before it is stored and is never returned by the API.</p>
      <form id="server-create-form" class="server-form">
        <label for="server-name">Display name</label>
        <input id="server-name" name="name" required maxlength="120" placeholder="Production host" />
        <div class="form-row">
          <div><label for="server-host">Host or IP address</label><input id="server-host" name="host" required placeholder="server.example.com" /></div>
          <div><label for="server-port">SSH port</label><input id="server-port" name="ssh_port" type="number" min="1" max="65535" value="22" required /></div>
        </div>
        <label for="server-username">SSH username</label>
        <input id="server-username" name="ssh_username" required placeholder="deploy" />
        <label for="server-private-key">Private key (OpenSSH format)</label>
        <textarea id="server-private-key" name="private_key" rows="5" required autocomplete="off" spellcheck="false" placeholder="-----BEGIN OPENSSH PRIVATE KEY-----"></textarea>
        <button type="submit" class="primary-button">Save server securely</button>
      </form>
    </section>
    <p id="server-feedback" class="server-feedback" role="status" aria-live="polite"></p>
    <div id="server-list" class="server-list"><article class="empty-card"><p>Loading server records…</p></article></div>
  `;

  const form = root.querySelector("#server-create-form");
  const feedback = root.querySelector("#server-feedback");
  const list = root.querySelector("#server-list");

  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    const submit = form.querySelector("button[type=submit]");
    const values = new FormData(form);
    submit.disabled = true;
    submit.textContent = "Encrypting and saving…";
    feedback.textContent = "";
    try {
      await apiFetch("/api/arsy/v1/servers", {
        method: "POST",
        body: JSON.stringify({
          name: String(values.get("name") || ""),
          host: String(values.get("host") || ""),
          ssh_port: Number(values.get("ssh_port")),
          ssh_username: String(values.get("ssh_username") || ""),
          private_key: String(values.get("private_key") || ""),
        }),
      });
      form.reset();
      feedback.textContent = "Server saved. Run the SSH test to inspect and confirm its host key.";
      await refreshServerList(list, feedback);
    } catch (error) {
      feedback.textContent = error.message || "Unable to save the server.";
    } finally {
      submit.disabled = false;
      submit.textContent = "Save server securely";
    }
  });

  list.addEventListener("click", async (event) => {
    const button = event.target.closest("button[data-action]");
    if (!button) return;
    const serverId = button.dataset.serverId;
    const action = button.dataset.action;
    if (!serverId) return;
    button.disabled = true;
    try {
      if (action === "test") {
        const result = await apiFetch(`/api/arsy/v1/servers/${encodeURIComponent(serverId)}/test`, { method: "POST" });
        if (result.code === "host_key_confirmation_required") {
          showHostKeyConfirmation(feedback, serverId, result.fingerprint);
        } else {
          feedback.textContent = result.message || "SSH test completed.";
        }
      } else if (action === "readiness") {
        try {
          const summary = await apiFetch(`/api/arsy/v1/servers/${encodeURIComponent(serverId)}/readiness`, { method: "POST" });
          showReadinessSummary(feedback, summary);
        } catch (error) {
          if (error.data && typeof error.data.ready === "boolean") showReadinessSummary(feedback, error.data);
          else throw error;
        }
      }
      await refreshServerList(list, feedback);
    } catch (error) {
      feedback.textContent = error.message || "The server operation failed.";
    } finally {
      button.disabled = false;
    }
  });

  void refreshServerList(list, feedback);
}

async function refreshServerList(list, feedback) {
  try {
    const data = await apiFetch("/api/arsy/v1/servers");
    const items = Array.isArray(data.items) ? data.items : [];
    if (!items.length) {
      list.innerHTML = '<article class="empty-card"><h2>No servers registered</h2><p>Register a Linux host above to begin the SSH trust and readiness workflow.</p></article>';
      return;
    }
    list.innerHTML = items.map((server) => `
      <article class="server-card">
        <div class="server-card-head">
          <div><h3>${escapeHTML(server.name)}</h3><p>${escapeHTML(server.host)}:${Number(server.ssh_port) || 22} · ${escapeHTML(server.ssh_username)}</p></div>
          <span class="server-state">${escapeHTML(server.readiness_state || "unknown")}</span>
        </div>
        <p class="server-meta">${escapeHTML([server.os_name, server.architecture].filter(Boolean).join(" · ") || "Readiness not checked")}</p>
        ${server.host_key_fingerprint ? `<p class="server-meta">Pinned key: <code>${escapeHTML(server.host_key_fingerprint)}</code></p>` : '<p class="server-meta">Host key not yet pinned</p>'}
        <div class="server-actions">
          <button class="quiet-button" data-action="test" data-server-id="${escapeHTML(server.id)}">Test SSH / host key</button>
          <button class="quiet-button" data-action="readiness" data-server-id="${escapeHTML(server.id)}" ${server.host_key_fingerprint ? "" : "disabled"}>Check readiness</button>
        </div>
      </article>
    `).join("");
  } catch (error) {
    list.innerHTML = `<article class="notice-card"><p>${escapeHTML(error.message || "Unable to load servers.")}</p></article>`;
    if (error.message?.includes("session")) feedback.textContent = error.message;
  }
}

function showHostKeyConfirmation(feedback, serverId, fingerprint) {
  feedback.innerHTML = `
    <article class="notice-card host-key-card">
      <h2>Verify this SSH host key</h2>
      <p>Compare this fingerprint with a trusted source on the target host before confirming. A-R-S-Y will not authenticate until you confirm the exact fingerprint.</p>
      <code class="fingerprint-value">${escapeHTML(fingerprint)}</code>
      <button id="confirm-host-key" class="primary-button">I verified this fingerprint</button>
    </article>
  `;
  feedback.querySelector("#confirm-host-key").addEventListener("click", async (event) => {
    const button = event.currentTarget;
    button.disabled = true;
    try {
      await apiFetch(`/api/arsy/v1/servers/${encodeURIComponent(serverId)}/host-key/confirm`, {
        method: "POST",
        body: JSON.stringify({ fingerprint, confirmed: true }),
      });
      feedback.textContent = "Fingerprint pinned. Testing SSH authentication…";
      const result = await apiFetch(`/api/arsy/v1/servers/${encodeURIComponent(serverId)}/test`, { method: "POST" });
      feedback.textContent = result.message || "SSH connection test completed.";
    } catch (error) {
      feedback.textContent = error.message || "Unable to confirm the host key.";
    }
  });
}

function showReadinessSummary(feedback, summary) {
  const names = ["ssh", "host_key", "os", "docker", "compose"];
  feedback.innerHTML = `
    <article class="readiness-card">
      <h2>${summary.ready ? "Server is ready" : "Readiness checks need attention"}</h2>
      <div class="readiness-list">
        ${names.map((name) => {
          const item = summary[name] || {};
          return `<div><strong>${escapeHTML(name.replace("_", " "))}</strong><span>${escapeHTML(item.code || "unknown")}</span><p>${escapeHTML(item.message || "")}${item.version ? " · " + escapeHTML(item.version) : ""}</p></div>`;
        }).join("")}
      </div>
    </article>
  `;
}

function escapeHTML(value) {
  return String(value ?? "").replace(/[&<>"']/g, (character) => ({
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    '"': "&quot;",
    "'": "&#39;",
  })[character]);
}
