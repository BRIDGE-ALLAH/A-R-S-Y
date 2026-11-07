const superuserAuthPath = '/api/collections/_superusers'

export async function getOnboardingStatus() {
  const response = await fetch('/api/arsy/v1/onboarding')
  return readResponse(response)
}

async function readResponse(response) {
  const result = await response.json().catch(() => ({}))
  if (!response.ok) {
    const error = new Error(result.message || 'PocketBase request failed.')
    error.status = response.status
    throw error
  }
  return result
}

export async function authenticateSuperuser(identity, password) {
  const response = await fetch(`${superuserAuthPath}/auth-with-password`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ identity, password }),
  })
  return readResponse(response)
}

export async function refreshSuperuser(token) {
  const response = await fetch(`${superuserAuthPath}/auth-refresh`, {
    method: 'POST',
    headers: { Authorization: token },
  })
  return readResponse(response)
}
