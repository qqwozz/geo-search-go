const BASE = '/api'

export async function search(query, lat, lon, { radius = 2000, limit = 20, sort = 'relevance' } = {}) {
  const params = new URLSearchParams({ q: query, lat, lon, radius, limit, sort })
  const res = await fetch(`${BASE}/search?${params}`)
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `Search failed (${res.status})`)
  }
  return res.json()
}

export async function autocomplete(query) {
  const params = query ? `?q=${encodeURIComponent(query)}` : ''
  const res = await fetch(`${BASE}/autocomplete${params}`)
  return res.json()
}

export async function health() {
  const res = await fetch(`${BASE}/health`)
  return res.json()
}
