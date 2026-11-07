let authToken = "";
let authRecord = null;

// The superuser token intentionally exists only in this module's memory.
export async function login(identity, password) {
  const response = await fetch("/api/collections/_superusers/auth-with-password", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ identity, password }),
  });
  const data = await response.json().catch(() => ({}));
  if (!response.ok || !data.token) {
    throw new Error(data.message || "Login failed. Check your credentials and try again.");
  }
  authToken = data.token;
  authRecord = data.record || null;
  return authRecord;
}

export function logout() {
  authToken = "";
  authRecord = null;
}

export function isAuthenticated() {
  return authToken.length > 0;
}

export function currentUser() {
  return authRecord;
}

export async function verifySession() {
  if (!authToken) return false;
  const response = await fetch("/api/arsy/v1/ready", {
    headers: { Authorization: authToken },
  });
  if (response.status === 401) {
    logout();
    return false;
  }
  return true;
}
