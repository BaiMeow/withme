async function request(url, options) {
  const resp = await fetch(url, options)
  const data = await resp.json().catch(() => ({}))
  if (!resp.ok) {
    throw new Error(data.error || `请求失败 (${resp.status})`)
  }
  if (data.error) {
    throw new Error(data.error)
  }
  return data
}

export function generateProfile(username, version) {
  return request('/api/generate', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, version }),
  })
}

export function getProfile(id) {
  return request(`/api/profiles/${encodeURIComponent(id)}`)
}

export function listProfiles() {
  return request('/api/profiles')
}
