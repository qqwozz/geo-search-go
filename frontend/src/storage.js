const MAX_HISTORY = 20

export function loadFromStorage(key, fallback) {
  try {
    const raw = localStorage.getItem(`geo_${key}`)
    return raw ? JSON.parse(raw) : fallback
  } catch {
    return fallback
  }
}

export function saveToStorage(key, value) {
  try {
    localStorage.setItem(`geo_${key}`, JSON.stringify(value))
  } catch {}
}

export function addHistory(history, entry) {
  const filtered = history.filter(h => h.query !== entry.query)
  return [entry, ...filtered].slice(0, MAX_HISTORY)
}
