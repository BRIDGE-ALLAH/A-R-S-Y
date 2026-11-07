import "./style.css";

const pages = {
  overview: {
    title: "Overview",
    description: "Your infrastructure at a glance.",
    heading: "Welcome to A-R-S-Y",
    body: "Register a server to begin managing your applications and deployments.",
  },
  servers: {
    title: "Servers",
    description: "Manage the Linux hosts connected to this control plane.",
    heading: "No servers registered",
    body: "Server registration will be available here after the SSH management flow is enabled.",
  },
  projects: {
    title: "Projects",
    description: "Group related applications into projects.",
    heading: "No projects yet",
    body: "Create a project after registering your first server.",
  },
  deployments: {
    title: "Deployments",
    description: "Review deployment activity and application status.",
    heading: "No deployments yet",
    body: "Deployment history will appear here after your first Compose deployment.",
  },
};

const nav = Object.entries(pages)
  .map(([id, page]) => `<a class="nav-link" href="#${id}" data-page="${id}">${page.title}</a>`)
  .join("");

document.querySelector("#app").innerHTML = `
  <div class="shell">
    <aside class="sidebar">
      <a class="brand" href="#overview"><span class="brand-mark">A</span><span>A-R-S-Y</span></a>
      <p class="nav-label">WORKSPACE</p>
      <nav aria-label="Main navigation">${nav}</nav>
      <div class="sidebar-foot"><span class="status-dot"></span> Control plane</div>
    </aside>
    <main class="main">
      <header class="topbar"><span>Self-hosted deployment control plane</span><span class="version">MVP · Early access</span></header>
      <section id="page" class="page" aria-live="polite"></section>
    </main>
  </div>
`;

function render() {
  const id = location.hash.slice(1) || "overview";
  const page = pages[id] || pages.overview;
  document.querySelectorAll(".nav-link").forEach((link) => {
    link.classList.toggle("active", link.dataset.page === id);
    if (link.dataset.page === id) link.setAttribute("aria-current", "page");
    else link.removeAttribute("aria-current");
  });
  document.querySelector("#page").innerHTML = `
    <div class="eyebrow">A-R-S-Y CONSOLE</div>
    <h1>${page.title}</h1>
    <p class="subtitle">${page.description}</p>
    <article class="empty-card">
      <div class="empty-icon" aria-hidden="true">◇</div>
      <h2>${page.heading}</h2>
      <p>${page.body}</p>
    </article>
  `;
}

window.addEventListener("hashchange", render);
render();
