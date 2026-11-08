(function(){let e=document.createElement(`link`).relList;if(e&&e.supports&&e.supports(`modulepreload`))return;for(let e of document.querySelectorAll(`link[rel="modulepreload"]`))n(e);new MutationObserver(e=>{for(let t of e)if(t.type===`childList`)for(let e of t.addedNodes)e.tagName===`LINK`&&e.rel===`modulepreload`&&n(e)}).observe(document,{childList:!0,subtree:!0});function t(e){let t={};return e.integrity&&(t.integrity=e.integrity),e.referrerPolicy&&(t.referrerPolicy=e.referrerPolicy),t.credentials=e.crossOrigin===`use-credentials`?`include`:e.crossOrigin===`anonymous`?`omit`:`same-origin`,t}function n(e){if(e.ep)return;e.ep=!0;let n=t(e);fetch(e.href,n)}})();var e=``,t=null;async function n(n,r){let i=await fetch(`/api/collections/_superusers/auth-with-password`,{method:`POST`,headers:{"Content-Type":`application/json`},body:JSON.stringify({identity:n,password:r})}),a=await i.json().catch(()=>({}));if(!i.ok||!a.token)throw Error(a.message||`Login failed. Check your credentials and try again.`);return e=a.token,t=a.record||null,t}function r(){e=``,t=null}function i(){return e.length>0}function a(){return t}async function o(){return e?(await fetch(`/api/arsy/v1/ready`,{headers:{Authorization:e}})).status!==401||(r(),!1):!1}async function s(t,n={}){if(!e)throw Error(`Your session has expired. Please sign in again.`);let i=new Headers(n.headers||{});i.set(`Authorization`,e),n.body&&!(n.body instanceof FormData)&&i.set(`Content-Type`,`application/json`);let a=await fetch(t,{...n,headers:i}),o=await a.json().catch(()=>({}));if(a.status===401)throw r(),Error(`Your session expired. Please sign in again.`);if(!a.ok){let e=Error(o.error||o.message||`Request failed (${a.status}).`);throw e.status=a.status,e.data=o,e}return o}var c={overview:[`Overview`,`Your infrastructure at a glance.`],servers:[`Servers`,`Manage the Linux hosts connected to this control plane.`],projects:[`Projects`,`Group related applications into projects.`],deployments:[`Deployments`,`Review deployment activity and application status.`],engine:[`Engine`,`Create tenants and manage engine records through the live API.`]},l=null,u=``,d=!1;async function f(){try{let e=await fetch(`/api/arsy/v1/onboarding`);if(!e.ok)throw Error(`The onboarding service is not available yet.`);l=await e.json()}catch(e){l={unavailable:!0,message:e.message}}_()}function p(){document.querySelector(`#app`).innerHTML=`
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
    </main>`}function m(){document.querySelector(`#app`).innerHTML=`
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
    </main>`;let e=document.querySelector(`#login-form`),t=document.querySelector(`#login-error`);t.textContent=u,e.addEventListener(`submit`,async r=>{r.preventDefault();let i=e.querySelector(`button`);i.disabled=!0,i.textContent=`Signing in…`,t.textContent=``;let a=new FormData(e);try{await n(String(a.get(`identity`)||``),String(a.get(`password`)||``)),u=``,d=!1,await _()}catch(e){u=e.message||`Unable to sign in.`,t.textContent=u,i.disabled=!1,i.textContent=`Sign in`}})}function h(){let e=Object.entries(c).map(([e,t])=>`<a class="nav-link" href="#${e}" data-page="${e}">${t[0]}</a>`).join(``),t=a();document.querySelector(`#app`).innerHTML=`
    <div class="shell">
      <aside class="sidebar">
        <a class="brand" href="#overview"><span class="brand-mark">A</span><span>A-R-S-Y</span></a>
        <p class="nav-label">WORKSPACE</p>
        <nav aria-label="Main navigation">${e}</nav>
        <div class="sidebar-foot"><span class="status-dot"></span> Control plane</div>
      </aside>
      <main class="main">
        <header class="topbar">
          <span>Self-hosted deployment control plane</span>
          <div class="topbar-actions"><span class="user-label">${t?.email||`Superuser`}</span><button id="logout" class="quiet-button">Sign out</button></div>
        </header>
        <section id="page" class="page" aria-live="polite"></section>
      </main>
    </div>`,document.querySelector(`#logout`).addEventListener(`click`,()=>{r(),u=``,_()}),document.querySelectorAll(`.nav-link`).forEach(e=>{e.addEventListener(`click`,()=>{setTimeout(g,0)})}),g()}function g(){let e=location.hash.slice(1)||`overview`,t=c[e]||c.overview;document.querySelectorAll(`.nav-link`).forEach(t=>{t.classList.toggle(`active`,t.dataset.page===e),t.dataset.page===e?t.setAttribute(`aria-current`,`page`):t.removeAttribute(`aria-current`)});let n=``;n=e===`overview`?`
      <article class="status-grid">
        <div class="status-card"><span>Superuser</span><strong>${l.superuser_configured?`Configured`:`Required`}</strong></div>
        <div class="status-card"><span>Secret key</span><strong>${l.secret_key_configured?`Configured`:`Action required`}</strong></div>
        <div class="status-card"><span>Managed servers</span><strong>${l.server_registered?`Registered`:`Not registered`}</strong></div>
      </article>
      ${l.secret_key_configured?``:`<article class="notice-card"><h2>Configure the secret key</h2><p>Set <code>ARSY_SECRET_KEY</code> in the control-plane process environment before storing SSH credentials. Keep the value outside source control.</p></article>`}
      ${l.server_registered?``:`<article class="empty-card"><h2>Register your first server</h2><p>Open Servers to continue setup. A-R-S-Y will require an SSH host-key fingerprint confirmation before trusting a host.</p></article>`}
    `:e===`servers`?`<div id="server-manager"><article class="empty-card"><h2>Loading servers…</h2></article></div>`:e===`projects`?`<div id="project-manager"><article class="empty-card"><h2>Loading projects…</h2></article></div>`:e===`engine`?`<div id="engine-manager"><article class="empty-card"><h2>Loading engine console…</h2></article></div>`:`<div id="deployment-manager"><article class="empty-card"><h2>Loading deployments…</h2></article></div>`,document.querySelector(`#page`).innerHTML=`
    <div class="eyebrow">A-R-S-Y CONSOLE</div>
    <h1>${t[0]}</h1>
    <p class="subtitle">${t[1]}</p>
    ${n}
  `,e===`servers`&&v(),e===`projects`&&T(),e===`deployments`&&E(),e===`engine`&&D()}async function _(){l?l.unavailable?(document.querySelector(`#app`).innerHTML=`<main class="auth-screen"><section class="auth-card"><h1>A-R-S-Y is starting</h1><p class="subtitle">${l.message}</p><button id="retry" class="primary-button">Retry</button></section></main>`,document.querySelector(`#retry`).addEventListener(`click`,f)):l.superuser_configured?i()?!d&&(d=!0,!await o())?(u=`Your session expired. Please sign in again.`,m()):h():m():p():document.querySelector(`#app`).innerHTML=`<main class="auth-screen"><section class="auth-card"><h1>Loading A-R-S-Y…</h1></section></main>`}window.addEventListener(`hashchange`,g),f();function v(){let e=document.querySelector(`#server-manager`);if(!e)return;e.innerHTML=`
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
  `;let t=e.querySelector(`#server-create-form`),n=e.querySelector(`#server-feedback`),r=e.querySelector(`#server-list`);t.addEventListener(`submit`,async e=>{e.preventDefault();let i=t.querySelector(`button[type=submit]`),a=new FormData(t);i.disabled=!0,i.textContent=`Encrypting and saving…`,n.textContent=``;try{await s(`/api/arsy/v1/servers`,{method:`POST`,body:JSON.stringify({name:String(a.get(`name`)||``),host:String(a.get(`host`)||``),ssh_port:Number(a.get(`ssh_port`)),ssh_username:String(a.get(`ssh_username`)||``),private_key:String(a.get(`private_key`)||``)})}),t.reset(),n.textContent=`Server saved. Run the SSH test to inspect and confirm its host key.`,await y(r,n)}catch(e){n.textContent=e.message||`Unable to save the server.`}finally{i.disabled=!1,i.textContent=`Save server securely`}}),r.addEventListener(`click`,async e=>{let t=e.target.closest(`button[data-action]`);if(!t)return;let i=t.dataset.serverId,a=t.dataset.action;if(i){t.disabled=!0;try{if(a===`test`){let e=await s(`/api/arsy/v1/servers/${encodeURIComponent(i)}/test`,{method:`POST`});e.code===`host_key_confirmation_required`?b(n,i,e.fingerprint):n.textContent=e.message||`SSH test completed.`}else if(a===`readiness`)try{let e=await s(`/api/arsy/v1/servers/${encodeURIComponent(i)}/readiness`,{method:`POST`});x(n,e)}catch(e){if(e.data&&typeof e.data.ready==`boolean`)x(n,e.data);else throw e}await y(r,n)}catch(e){n.textContent=e.message||`The server operation failed.`}finally{t.disabled=!1}}}),y(r,n)}async function y(e,t){try{let t=await s(`/api/arsy/v1/servers`),n=Array.isArray(t.items)?t.items:[];if(!n.length){e.innerHTML=`<article class="empty-card"><h2>No servers registered</h2><p>Register a Linux host above to begin the SSH trust and readiness workflow.</p></article>`;return}e.innerHTML=n.map(e=>`
      <article class="server-card">
        <div class="server-card-head">
          <div><h3>${S(e.name)}</h3><p>${S(e.host)}:${Number(e.ssh_port)||22} · ${S(e.ssh_username)}</p></div>
          <span class="server-state">${S(e.readiness_state||`unknown`)}</span>
        </div>
        <p class="server-meta">${S([e.os_name,e.architecture].filter(Boolean).join(` · `)||`Readiness not checked`)}</p>
        ${e.host_key_fingerprint?`<p class="server-meta">Pinned key: <code>${S(e.host_key_fingerprint)}</code></p>`:`<p class="server-meta">Host key not yet pinned</p>`}
        <div class="server-actions">
          <button class="quiet-button" data-action="test" data-server-id="${S(e.id)}">Test SSH / host key</button>
          <button class="quiet-button" data-action="readiness" data-server-id="${S(e.id)}" ${e.host_key_fingerprint?``:`disabled`}>Check readiness</button>
        </div>
      </article>
    `).join(``)}catch(n){e.innerHTML=`<article class="notice-card"><p>${S(n.message||`Unable to load servers.`)}</p></article>`,n.message?.includes(`session`)&&(t.textContent=n.message)}}function b(e,t,n){e.innerHTML=`
    <article class="notice-card host-key-card">
      <h2>Verify this SSH host key</h2>
      <p>Compare this fingerprint with a trusted source on the target host before confirming. A-R-S-Y will not authenticate until you confirm the exact fingerprint.</p>
      <code class="fingerprint-value">${S(n)}</code>
      <button id="confirm-host-key" class="primary-button">I verified this fingerprint</button>
    </article>
  `,e.querySelector(`#confirm-host-key`).addEventListener(`click`,async r=>{let i=r.currentTarget;i.disabled=!0;try{await s(`/api/arsy/v1/servers/${encodeURIComponent(t)}/host-key/confirm`,{method:`POST`,body:JSON.stringify({fingerprint:n,confirmed:!0})}),e.textContent=`Fingerprint pinned. Testing SSH authentication…`,e.textContent=(await s(`/api/arsy/v1/servers/${encodeURIComponent(t)}/test`,{method:`POST`})).message||`SSH connection test completed.`}catch(t){e.textContent=t.message||`Unable to confirm the host key.`}})}function x(e,t){e.innerHTML=`
    <article class="readiness-card">
      <h2>${t.ready?`Server is ready`:`Readiness checks need attention`}</h2>
      <div class="readiness-list">
        ${[`ssh`,`host_key`,`os`,`docker`,`compose`].map(e=>{let n=t[e]||{};return`<div><strong>${S(e.replace(`_`,` `))}</strong><span>${S(n.code||`unknown`)}</span><p>${S(n.message||``)}${n.version?` · `+S(n.version):``}</p></div>`}).join(``)}
      </div>
    </article>
  `}function S(e){return String(e??``).replace(/[&<>"']/g,e=>({"&":`&amp;`,"<":`&lt;`,">":`&gt;`,'"':`&quot;`,"'":`&#39;`})[e])}var C=new Set([`queued`,`preparing`,`uploading`,`validating`,`deploying`,`verifying`]);async function w(e,t){let n=0;for(;n++<200;){try{let n=await s(`/api/arsy/v1/deployments/`+encodeURIComponent(e)+`/progress`),r=typeof n.percent==`number`?n.percent:0,i=S(n.step||n.state||`Running`);if(t&&(t.innerHTML=`<div class="progress-bar-wrap"><div class="progress-bar" style="width:${r}%"></div></div><p class="progress-step">${i} (${r}%)</p>`),!C.has(n.state))break}catch{}await new Promise(e=>setTimeout(e,3e3))}}async function T(){let e=document.querySelector(`#project-manager`);if(!e)return;e.innerHTML=`
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
        <textarea id="app-manifest" name="compose_manifest" rows="6" required placeholder="services:
  web:
    image: nginx:alpine
    ports:
      - "8080:80""></textarea>
        <button type="submit" class="primary-button">Create application</button>
      </form>
    </section>
    <p id="project-feedback" class="server-feedback" role="status"></p>
    <div id="project-list" class="server-list"></div>
  `;let t=e.querySelector(`#project-create-form`),n=e.querySelector(`#app-create-form`),r=e.querySelector(`#project-feedback`),i=e.querySelector(`#project-list`);async function a(){try{let[e,t]=await Promise.all([s(`/api/arsy/v1/projects`),s(`/api/arsy/v1/servers`)]),r=n.querySelector(`#app-project`),i=n.querySelector(`#app-server`);r.innerHTML=(e.items||[]).map(e=>`<option value="${S(e.id)}">${S(e.name)}</option>`).join(``),i.innerHTML=(t.items||[]).map(e=>`<option value="${S(e.id)}">${S(e.name)} (${S(e.host)})</option>`).join(``)}catch{}}async function o(){try{let[e,t]=await Promise.all([s(`/api/arsy/v1/projects`),s(`/api/arsy/v1/applications`)]),n=e.items||[],r=t.items||[];if(!n.length){i.innerHTML=`<article class="empty-card"><h2>No projects yet</h2><p>Create your first project above to group applications.</p></article>`;return}i.innerHTML=n.map(e=>{let t=r.filter(t=>t.project_id===e.id);return`
          <article class="project-card">
            <h3>${S(e.name)}</h3>
            <p>${S(e.description||`No description`)}</p>
            <div class="card-meta"><span>Slug: <code>${S(e.slug)}</code></span><span>Applications: ${t.length}</span></div>
            ${t.length?`
              <div style="margin-top:14px; border-top:1px solid #29364a; padding-top:10px;">
                <strong>Applications in project:</strong>
                ${t.map(e=>`
                  <div class="app-card" style="margin-top:8px;">
                    <div style="display:flex; justify-content:space-between; align-items:center;">
                      <div><strong>${S(e.name)}</strong> · <small>${S(e.slug)}</small></div>
                      <div style="display:flex;align-items:center;gap:8px;">
                        <span class="deploy-status-tag ${S(e.lifecycle_state)}">${S(e.lifecycle_state)}</span>
                        <button class="deploy-btn quiet-button" data-action="deploy-app" data-app-id="${S(e.id)}" data-app-name="${S(e.name)}">Deploy</button>
                      </div>
                    </div>
                    <div id="deploy-progress-${S(e.id)}" class="deploy-progress" style="display:none;"></div>
                  </div>
                `).join(``)}
              </div>
            `:``}
          </article>
        `}).join(``)}catch(e){i.innerHTML=`<article class="notice-card"><p>${S(e.message)}</p></article>`}}t.addEventListener(`submit`,async e=>{e.preventDefault();let n=t.querySelector(`button`);n.disabled=!0,r.textContent=``;let i=new FormData(t);try{await s(`/api/arsy/v1/projects`,{method:`POST`,body:JSON.stringify({name:i.get(`name`),description:i.get(`description`)})}),t.reset(),r.textContent=`Project created successfully.`,await Promise.all([o(),a()])}catch(e){r.textContent=e.message||`Failed to create project.`}finally{n.disabled=!1}}),n.addEventListener(`submit`,async e=>{e.preventDefault();let t=n.querySelector(`button`);t.disabled=!0,r.textContent=``;let i=new FormData(n);try{await s(`/api/arsy/v1/applications`,{method:`POST`,body:JSON.stringify({project_id:i.get(`project_id`),server_id:i.get(`server_id`),name:i.get(`name`),compose_manifest:i.get(`compose_manifest`)})}),n.reset(),r.textContent=`Application created successfully.`,await o()}catch(e){r.textContent=e.message||`Failed to create application.`}finally{t.disabled=!1}}),i.addEventListener(`click`,async e=>{let t=e.target.closest(`button[data-action='deploy-app']`);if(!t)return;let n=t.dataset.appId,i=t.dataset.appName;t.disabled=!0,r.textContent=``;let a=document.getElementById(`deploy-progress-`+n);a&&(a.style.display=`block`,a.innerHTML=`<div class="progress-bar-wrap"><div class="progress-bar" style="width:0%"></div></div><p class="progress-step">Queuing deployment…</p>`);try{let e=(await s(`/api/arsy/v1/deployments`,{method:`POST`,body:JSON.stringify({application_id:n})})).deployment_id;r.textContent=i+` deployment queued (`+e.slice(0,8)+`…)`,await w(e,a)}catch(e){r.textContent=e.message||`Deployment failed.`,a&&(a.innerHTML=`<p class="progress-error">`+S(e.message||`Failed`)+`</p>`)}finally{t.disabled=!1}}),a(),o()}async function E(){let e=document.querySelector(`#deployment-manager`);if(!e)return;e.innerHTML=`
    <div id="deployment-feedback" class="server-feedback" role="status"></div>
    <div id="deployment-list" class="server-list"><article class="empty-card"><p>Loading deployments…</p></article></div>
  `;let t=e.querySelector(`#deployment-list`),n=e.querySelector(`#deployment-feedback`);async function r(){try{let e=await s(`/api/arsy/v1/deployments`),n=Array.isArray(e)?e:e.items||[];if(!n||!n.length){t.innerHTML=`<article class="empty-card"><h2>No deployments yet</h2><p>Deployments will appear here once applications are deployed.</p></article>`;return}t.innerHTML=n.map(e=>`
        <article class="deploy-card">
          <div style="display:flex; justify-content:space-between; align-items:flex-start;">
            <div>
              <h3>Deployment <code>${S(e.deployment_id.slice(0,8))}</code></h3>
              <p>Application: <code>${S(e.application_id)}</code></p>
            </div>
            <span class="deploy-status-tag ${S(e.state)}">${S(e.state)}</span>
          </div>
          <div class="card-meta">
          ${C.has(e.state)?`<div class="progress-bar-wrap" style="margin:8px 0;"><div class="progress-bar progress-bar-active"></div></div><p class="progress-step">In progress…</p>`:``}
            ${e.started_at?`<span>Started: ${S(e.started_at)}</span>`:``}
            ${e.finished_at?`<span>Finished: ${S(e.finished_at)}</span>`:``}
          </div>
          <div style="margin-top:12px; display:flex; gap:8px;">
            <button class="quiet-button" data-action="view-log" data-id="${S(e.deployment_id)}">View Logs</button>
            <button class="quiet-button" data-action="stop" data-id="${S(e.deployment_id)}" ${C.has(e.state)?``:`disabled`}>Stop</button>
            <button class="quiet-button" data-action="redeploy" data-id="${S(e.deployment_id)}" ${C.has(e.state)?`disabled`:``}>Redeploy</button>
            <button class="quiet-button" data-action="rollback" data-id="${S(e.deployment_id)}" ${e.state===`succeeded`&&e.previous_deployment_id?``:`disabled`}>Rollback</button>
          </div>
          <div id="log-${S(e.deployment_id)}" style="display:none;" class="log-output"></div>
        </article>
      `).join(``)}catch(e){t.innerHTML=`<article class="notice-card"><p>${S(e.message)}</p></article>`}}t.addEventListener(`click`,async t=>{let i=t.target.closest(`button[data-action]`);if(!i)return;let a=i.dataset.action,o=i.dataset.id;i.disabled=!0;try{if(a===`view-log`){let t=e.querySelector(`#log-${o}`);t.style.display===`block`?t.style.display=`none`:(t.style.display=`block`,t.textContent=`Fetching logs…`,t.textContent=(await s(`/api/arsy/v1/deployments/${o}/logs`)).log||`(empty)`)}else a===`stop`?(await s(`/api/arsy/v1/deployments/${o}/stop`,{method:`POST`}),n.textContent=`Deployment ${o.slice(0,8)} cancelled/stopped.`,await r()):a===`redeploy`?(await s(`/api/arsy/v1/deployments/${o}/redeploy`,{method:`POST`}),n.textContent=`Redeploy requested for ${o.slice(0,8)}.`):a===`rollback`&&(await s(`/api/arsy/v1/deployments/${o}/rollback`,{method:`POST`}),n.textContent=`Rollback requested for ${o.slice(0,8)}.`)}catch(e){n.textContent=e.message||`Operation failed.`}finally{i.disabled=!1}}),r()}async function D(){let e=document.querySelector(`#engine-manager`);if(!e)return;e.innerHTML=`
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
  `;let t=e.querySelector(`#engine-feedback`),n=e.querySelector(`#engine-list`);async function r(){try{let e=await s(`/api/arsy/v1/tenants`),t=Array.isArray(e.items)?e.items:[];n.innerHTML=t.length?t.map(e=>`
        <article class="server-card">
          <div class="server-card-head"><div><h3>${S(e.display_name||e.slug)}</h3><p><code>${S(e.slug)}</code> · ${S(e.status||`unknown`)}</p></div></div>
          <div class="server-actions"><button class="quiet-button" data-action="provision-tenant" data-slug="${S(e.slug)}">Provision schema</button></div>
        </article>`).join(``):`<article class="empty-card"><h2>No tenants yet</h2><p>Create a tenant to use the engine API.</p></article>`}catch(e){n.innerHTML=`<article class="notice-card"><p>${S(e.message||`Unable to load tenants.`)}</p></article>`}}e.querySelector(`#tenant-create-form`).addEventListener(`submit`,async e=>{e.preventDefault();let n=e.target.querySelector(`button`);n.disabled=!0,t.textContent=``;let i=new FormData(e.target);try{await s(`/api/arsy/v1/tenants`,{method:`POST`,body:JSON.stringify({slug:i.get(`slug`),display_name:i.get(`display_name`)})}),e.target.reset(),t.textContent=`Tenant created and provisioned.`,await r()}catch(e){t.textContent=e.message||`Failed to create tenant.`}finally{n.disabled=!1}}),e.querySelector(`#engine-record-form`).addEventListener(`submit`,async e=>{e.preventDefault();let n=e.target.querySelector(`button`);n.disabled=!0,t.textContent=``;let r=new FormData(e.target),i=String(r.get(`tenant`)||``).trim();try{await s(`/api/arsy/v1/engine/base.company/records`,{method:`POST`,headers:{"X-Arsy-Tenant":i},body:JSON.stringify({values:{id:r.get(`id`),name:r.get(`name`),code:r.get(`code`)},allowed_fields:[`id`,`name`,`code`]})}),t.textContent=`Company record saved for tenant `+i+`.`}catch(e){t.textContent=e.message||`Failed to save engine record.`}finally{n.disabled=!1}}),n.addEventListener(`click`,async e=>{let n=e.target.closest(`button[data-action='provision-tenant']`);if(n){n.disabled=!0,t.textContent=``;try{await s(`/api/arsy/v1/tenants/`+encodeURIComponent(n.dataset.slug)+`/provision`,{method:`POST`}),t.textContent=`Tenant schema provisioned.`,await r()}catch(e){t.textContent=e.message||`Provisioning failed.`}finally{n.disabled=!1}}}),await r()}