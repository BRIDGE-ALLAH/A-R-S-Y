import "./style.css";
import { apiFetch, currentUser, isAuthenticated, login, logout, verifySession } from "./auth.js";

const pages = {
  overview: ["Overview", "Your infrastructure at a glance."],
  servers: ["Servers", "Manage the Linux hosts connected to this control plane."],
  projects: ["Projects", "Group related applications into projects."],
  deployments: ["Deployments", "Review deployment activity and application status."],
  engine: ["Engine", "Create tenants and manage engine records through the live API."],
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
    content = '<div id="project-manager"><article class="empty-card"><h2>Loading projects…</h2></article></div>';
  } else if (id === "engine") {
    content = '<div id="engine-manager"><article class="empty-card"><h2>Loading engine console…</h2></article></div>';
  } else {
    content = '<div id="deployment-manager"><article class="empty-card"><h2>Loading deployments…</h2></article></div>';
  }

  document.querySelector("#page").innerHTML = `
    <div class="eyebrow">A-R-S-Y CONSOLE</div>
    <h1>${page[0]}</h1>
    <p class="subtitle">${page[1]}</p>
    ${content}
  `;
  if (id === "servers") void renderServerManager();
  if (id === "projects") void renderProjectManager();
  if (id === "deployments") void renderDeploymentManager();
  if (id === "engine") void renderEngineManager();
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


const ACTIVE_STATES = new Set(["queued", "preparing", "uploading", "validating", "deploying", "verifying"]);

async function pollDeploymentProgress(deploymentId, progressEl) {
  const MAX_POLLS = 200; // 200 * 3s = 10 minutes max
  let polls = 0;
  while (polls++ < MAX_POLLS) {
    try {
      const prog = await apiFetch("/api/arsy/v1/deployments/" + encodeURIComponent(deploymentId) + "/progress");
      const pct = typeof prog.percent === "number" ? prog.percent : 0;
      const step = escapeHTML(prog.step || prog.state || "Running");
      if (progressEl) {
        progressEl.innerHTML = `<div class="progress-bar-wrap"><div class="progress-bar" style="width:${pct}%"></div></div><p class="progress-step">${step} (${pct}%)</p>`;
      }
      if (!ACTIVE_STATES.has(prog.state)) break; // terminal state
    } catch (_) {}
    await new Promise(r => setTimeout(r, 3000)); // bounded 3s poll
  }
}
async function renderProjectManager() {
  const root = document.querySelector("#project-manager");
  if (!root) return;
  root.innerHTML = `
    <section class="form-card">
      <h2>Create a project</h2>
      <form id="project-create-form">
        <label for="project-name">Project name</label>
        <input id="project-name" name="name" required placeholder="Production Platform" />
        <label for="project-desc">Description</label>
        <input id="project-desc" name="description" placeholder="Main microservices stack" />
        <button type="submit" class="primary-button">Create project</button>
      </form>
    </section>
    <section class="form-card">
      <h2>Create an application</h2>
      <form id="app-create-form">
        <div class="form-row">
          <div><label for="app-project">Project</label><select id="app-project" name="project_id" required></select></div>
          <div><label for="app-server">Server</label><select id="app-server" name="server_id" required></select></div>
        </div>
        <label for="app-name">Application name</label>
        <input id="app-name" name="name" required placeholder="Web API" />
        <label for="app-manifest">Docker Compose manifest</label>
        <textarea id="app-manifest" name="compose_manifest" rows="6" required placeholder="services:\n  web:\n    image: nginx:alpine\n    ports:\n      - \"8080:80\""></textarea>
        <button type="submit" class="primary-button">Create application</button>
      </form>
    </section>
    <p id="project-feedback" class="server-feedback" role="status"></p>
    <div id="project-list" class="server-list"></div>
  `;

  const projectForm = root.querySelector("#project-create-form");
  const appForm = root.querySelector("#app-create-form");
  const feedback = root.querySelector("#project-feedback");
  const list = root.querySelector("#project-list");

  async function loadSelectOptions() {
    try {
      const [projData, servData] = await Promise.all([
        apiFetch("/api/arsy/v1/projects"),
        apiFetch("/api/arsy/v1/servers")
      ]);
      const projSelect = appForm.querySelector("#app-project");
      const servSelect = appForm.querySelector("#app-server");
      projSelect.innerHTML = (projData.items || []).map(p => `<option value="${escapeHTML(p.id)}">${escapeHTML(p.name)}</option>`).join("");
      servSelect.innerHTML = (servData.items || []).map(s => `<option value="${escapeHTML(s.id)}">${escapeHTML(s.name)} (${escapeHTML(s.host)})</option>`).join("");
    } catch (_) {}
  }

  async function refreshProjects() {
    try {
      const [projData, appData] = await Promise.all([
        apiFetch("/api/arsy/v1/projects"),
        apiFetch("/api/arsy/v1/applications")
      ]);
      const projects = projData.items || [];
      const apps = appData.items || [];
      if (!projects.length) {
        list.innerHTML = '<article class="empty-card"><h2>No projects yet</h2><p>Create your first project above to group applications.</p></article>';
        return;
      }
      list.innerHTML = projects.map(p => {
        const pApps = apps.filter(a => a.project_id === p.id);
        return `
          <article class="project-card">
            <h3>${escapeHTML(p.name)}</h3>
            <p>${escapeHTML(p.description || "No description")}</p>
            <div class="card-meta"><span>Slug: <code>${escapeHTML(p.slug)}</code></span><span>Applications: ${pApps.length}</span></div>
            ${pApps.length ? `
              <div style="margin-top:14px; border-top:1px solid #29364a; padding-top:10px;">
                <strong>Applications in project:</strong>
                ${pApps.map(a => `
                  <div class="app-card" style="margin-top:8px;">
                    <div style="display:flex; justify-content:space-between; align-items:center;">
                      <div><strong>${escapeHTML(a.name)}</strong> · <small>${escapeHTML(a.slug)}</small></div>
                      <div style="display:flex;align-items:center;gap:8px;">
                        <span class="deploy-status-tag ${escapeHTML(a.lifecycle_state)}">${escapeHTML(a.lifecycle_state)}</span>
                        <button class="deploy-btn quiet-button" data-action="deploy-app" data-app-id="${escapeHTML(a.id)}" data-app-name="${escapeHTML(a.name)}">Deploy</button>
                      </div>
                    </div>
                    <div id="deploy-progress-${escapeHTML(a.id)}" class="deploy-progress" style="display:none;"></div>
                  </div>
                `).join("")}
              </div>
            ` : ""}
          </article>
        `;
      }).join("");
    } catch (err) {
      list.innerHTML = `<article class="notice-card"><p>${escapeHTML(err.message)}</p></article>`;
    }
  }

  projectForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    const btn = projectForm.querySelector("button");
    btn.disabled = true;
    feedback.textContent = "";
    const data = new FormData(projectForm);
    try {
      await apiFetch("/api/arsy/v1/projects", {
        method: "POST",
        body: JSON.stringify({ name: data.get("name"), description: data.get("description") })
      });
      projectForm.reset();
      feedback.textContent = "Project created successfully.";
      await Promise.all([refreshProjects(), loadSelectOptions()]);
    } catch (err) {
      feedback.textContent = err.message || "Failed to create project.";
    } finally {
      btn.disabled = false;
    }
  });

  appForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    const btn = appForm.querySelector("button");
    btn.disabled = true;
    feedback.textContent = "";
    const data = new FormData(appForm);
    try {
      await apiFetch("/api/arsy/v1/applications", {
        method: "POST",
        body: JSON.stringify({
          project_id: data.get("project_id"),
          server_id: data.get("server_id"),
          name: data.get("name"),
          compose_manifest: data.get("compose_manifest")
        })
      });
      appForm.reset();
      feedback.textContent = "Application created successfully.";
      await refreshProjects();
    } catch (err) {
      feedback.textContent = err.message || "Failed to create application.";
    } finally {
      btn.disabled = false;
    }
  });


  list.addEventListener("click", async (e) => {
    const btn = e.target.closest("button[data-action='deploy-app']");
    if (!btn) return;
    const appId = btn.dataset.appId;
    const appName = btn.dataset.appName;
    btn.disabled = true;
    feedback.textContent = "";
    const progressEl = document.getElementById("deploy-progress-" + appId);
    if (progressEl) { progressEl.style.display = "block"; progressEl.innerHTML = '<div class="progress-bar-wrap"><div class="progress-bar" style="width:0%"></div></div><p class="progress-step">Queuing deployment…</p>'; }
    try {
      const result = await apiFetch("/api/arsy/v1/deployments", {
        method: "POST",
        body: JSON.stringify({ application_id: appId })
      });
      const deployId = result.deployment_id;
      feedback.textContent = appName + " deployment queued (" + deployId.slice(0, 8) + "…)";
      await pollDeploymentProgress(deployId, progressEl);
    } catch (err) {
      feedback.textContent = err.message || "Deployment failed.";
      if (progressEl) { progressEl.innerHTML = '<p class="progress-error">' + escapeHTML(err.message || "Failed") + '</p>'; }
    } finally {
      btn.disabled = false;
    }
  });
  void loadSelectOptions();
  void refreshProjects();
}

async function renderDeploymentManager() {
  const root = document.querySelector("#deployment-manager");
  if (!root) return;
  root.innerHTML = `
    <div id="deployment-feedback" class="server-feedback" role="status"></div>
    <div id="deployment-list" class="server-list"><article class="empty-card"><p>Loading deployments…</p></article></div>
  `;

  const list = root.querySelector("#deployment-list");
  const feedback = root.querySelector("#deployment-feedback");

  async function refreshDeployments() {
    try {
      const payload = await apiFetch("/api/arsy/v1/deployments");
      const deployments = Array.isArray(payload) ? payload : (payload.items || []);
      if (!deployments || !deployments.length) {
        list.innerHTML = '<article class="empty-card"><h2>No deployments yet</h2><p>Deployments will appear here once applications are deployed.</p></article>';
        return;
      }
      list.innerHTML = deployments.map(d => `
        <article class="deploy-card">
          <div style="display:flex; justify-content:space-between; align-items:flex-start;">
            <div>
              <h3>Deployment <code>${escapeHTML(d.deployment_id.slice(0, 8))}</code></h3>
              <p>Application: <code>${escapeHTML(d.application_id)}</code></p>
            </div>
            <span class="deploy-status-tag ${escapeHTML(d.state)}">${escapeHTML(d.state)}</span>
          </div>
          <div class="card-meta">
          ${ACTIVE_STATES.has(d.state) ? `<div class="progress-bar-wrap" style="margin:8px 0;"><div class="progress-bar progress-bar-active"></div></div><p class="progress-step">In progress…</p>` : ""}
            ${d.started_at ? `<span>Started: ${escapeHTML(d.started_at)}</span>` : ""}
            ${d.finished_at ? `<span>Finished: ${escapeHTML(d.finished_at)}</span>` : ""}
          </div>
          <div style="margin-top:12px; display:flex; gap:8px;">
            <button class="quiet-button" data-action="view-log" data-id="${escapeHTML(d.deployment_id)}">View Logs</button>
            <button class="quiet-button" data-action="stop" data-id="${escapeHTML(d.deployment_id)}" ${ACTIVE_STATES.has(d.state) ? "" : "disabled"}>Stop</button>
            <button class="quiet-button" data-action="redeploy" data-id="${escapeHTML(d.deployment_id)}" ${ACTIVE_STATES.has(d.state) ? "disabled" : ""}>Redeploy</button>
            <button class="quiet-button" data-action="rollback" data-id="${escapeHTML(d.deployment_id)}" ${d.state === "succeeded" && d.previous_deployment_id ? "" : "disabled"}>Rollback</button>
          </div>
          <div id="log-${escapeHTML(d.deployment_id)}" style="display:none;" class="log-output"></div>
        </article>
      `).join("");
    } catch (err) {
      list.innerHTML = `<article class="notice-card"><p>${escapeHTML(err.message)}</p></article>`;
    }
  }

  list.addEventListener("click", async (e) => {
    const btn = e.target.closest("button[data-action]");
    if (!btn) return;
    const action = btn.dataset.action;
    const id = btn.dataset.id;
    btn.disabled = true;
    try {
      if (action === "view-log") {
        const logBox = root.querySelector(`#log-${id}`);
        if (logBox.style.display === "block") {
          logBox.style.display = "none";
        } else {
          logBox.style.display = "block";
          logBox.textContent = "Fetching logs…";
          const res = await apiFetch(`/api/arsy/v1/deployments/${id}/logs`);
          logBox.textContent = res.log || "(empty)";
        }
      } else if (action === "stop") {
        await apiFetch(`/api/arsy/v1/deployments/${id}/stop`, { method: "POST" });
        feedback.textContent = `Deployment ${id.slice(0, 8)} cancelled/stopped.`;
        await refreshDeployments();
      } else if (action === "redeploy") {
        await apiFetch(`/api/arsy/v1/deployments/${id}/redeploy`, { method: "POST" });
        feedback.textContent = `Redeploy requested for ${id.slice(0, 8)}.`;
      } else if (action === "rollback") {
        await apiFetch(`/api/arsy/v1/deployments/${id}/rollback`, { method: "POST" });
        feedback.textContent = `Rollback requested for ${id.slice(0, 8)}.`;
      }
    } catch (err) {
      feedback.textContent = err.message || "Operation failed.";
    } finally {
      btn.disabled = false;
    }
  });

  void refreshDeployments();
}


async function renderEngineManager() {
  const root = document.querySelector("#engine-manager");
  if (!root) return;
  root.innerHTML = `
    <section class="form-card">
      <h2>Create a tenant</h2>
      <p>Tenants use isolated databases. Creating a tenant provisions the starter engine schema.</p>
      <form id="tenant-create-form">
        <label for="tenant-slug">Tenant slug</label>
        <input id="tenant-slug" name="slug" required maxlength="63" placeholder="acme-co" />
        <label for="tenant-name">Display name</label>
        <input id="tenant-name" name="display_name" required maxlength="120" placeholder="Acme Co" />
        <button type="submit" class="primary-button">Create tenant</button>
      </form>
    </section>
    <section class="form-card">
      <h2>Engine records (base.company)</h2>
      <p>This uses the live tenant-scoped engine API. Writes require explicit allowed fields.</p>
      <form id="engine-record-form">
        <label for="engine-tenant">Tenant slug</label>
        <input id="engine-tenant" name="tenant" required placeholder="acme-co" />
        <div class="form-row"><div><label for="company-id">Company ID</label><input id="company-id" name="id" required placeholder="c1" /></div><div><label for="company-code">Code</label><input id="company-code" name="code" placeholder="ACME" /></div></div>
        <label for="company-name">Name</label>
        <input id="company-name" name="name" required placeholder="Acme" />
        <button type="submit" class="primary-button">Save company record</button>
      </form>
    </section>
    <p id="engine-feedback" class="server-feedback" role="status"></p>
    <div id="engine-list" class="server-list"><article class="empty-card"><p>Loading tenants…</p></article></div>
  `;

  const feedback = root.querySelector("#engine-feedback");
  const list = root.querySelector("#engine-list");

  async function refreshTenants() {
    try {
      const data = await apiFetch("/api/arsy/v1/tenants");
      const items = Array.isArray(data.items) ? data.items : [];
      list.innerHTML = items.length ? items.map((t) => `
        <article class="server-card">
          <div class="server-card-head"><div><h3>${escapeHTML(t.display_name || t.slug)}</h3><p><code>${escapeHTML(t.slug)}</code> · ${escapeHTML(t.status || "unknown")}</p></div></div>
          <div class="server-actions"><button class="quiet-button" data-action="provision-tenant" data-slug="${escapeHTML(t.slug)}">Provision schema</button></div>
        </article>`).join("") : '<article class="empty-card"><h2>No tenants yet</h2><p>Create a tenant to use the engine API.</p></article>';
    } catch (err) { list.innerHTML = `<article class="notice-card"><p>${escapeHTML(err.message || "Unable to load tenants.")}</p></article>`; }
  }

  root.querySelector("#tenant-create-form").addEventListener("submit", async (e) => {
    e.preventDefault();
    const btn = e.target.querySelector("button");
    btn.disabled = true; feedback.textContent = "";
    const values = new FormData(e.target);
    try {
      await apiFetch("/api/arsy/v1/tenants", { method: "POST", body: JSON.stringify({ slug: values.get("slug"), display_name: values.get("display_name") }) });
      e.target.reset(); feedback.textContent = "Tenant created and provisioned."; await refreshTenants();
    } catch (err) { feedback.textContent = err.message || "Failed to create tenant."; } finally { btn.disabled = false; }
  });

  root.querySelector("#engine-record-form").addEventListener("submit", async (e) => {
    e.preventDefault();
    const btn = e.target.querySelector("button");
    btn.disabled = true; feedback.textContent = "";
    const values = new FormData(e.target);
    const tenant = String(values.get("tenant") || "").trim();
    try {
      await apiFetch("/api/arsy/v1/engine/base.company/records", { method: "POST", headers: { "X-Arsy-Tenant": tenant }, body: JSON.stringify({ values: { id: values.get("id"), name: values.get("name"), code: values.get("code") }, allowed_fields: ["id","name","code"] }) });
      feedback.textContent = "Company record saved for tenant " + tenant + ".";
    } catch (err) { feedback.textContent = err.message || "Failed to save engine record."; } finally { btn.disabled = false; }
  });

  list.addEventListener("click", async (e) => {
    const btn = e.target.closest("button[data-action='provision-tenant']");
    if (!btn) return;
    btn.disabled = true; feedback.textContent = "";
    try { await apiFetch("/api/arsy/v1/tenants/" + encodeURIComponent(btn.dataset.slug) + "/provision", { method: "POST" }); feedback.textContent = "Tenant schema provisioned."; await refreshTenants(); }
    catch (err) { feedback.textContent = err.message || "Provisioning failed."; }
    finally { btn.disabled = false; }
  });

  await refreshTenants();
}
