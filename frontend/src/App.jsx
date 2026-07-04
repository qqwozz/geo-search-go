import { useState, useEffect, useCallback } from 'react'
import { search } from './api'
import { loadFromStorage, saveToStorage, addHistory } from './storage'
import Header from './components/Header'
import SearchBar from './components/SearchBar'
import QuickFilters from './components/QuickFilters'
import ResultsList from './components/ResultsList'
import MapView from './components/MapView'
import DetailModal from './components/DetailModal'
import './App.css'

const DEFAULT_CENTER = { lat: 55.755, lon: 37.615 }

function parseUrlParams() {
  const p = new URLSearchParams(window.location.search)
  const q = p.get('q')
  const lat = parseFloat(p.get('lat'))
  const lon = parseFloat(p.get('lon'))
  const radius = parseInt(p.get('radius')) || undefined
  if (q && !isNaN(lat) && !isNaN(lon)) return { q, lat, lon, radius }
  return null
}

export default function App() {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState([])
  const [total, setTotal] = useState(0)
  const [cached, setCached] = useState(false)
  const [center, setCenter] = useState(DEFAULT_CENTER)
  const [userPosition, setUserPosition] = useState(null)
  const [selectedPOI, setSelectedPOI] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [favorites, setFavorites] = useState(() => loadFromStorage('favorites', []))
  const [history, setHistory] = useState(() => loadFromStorage('history', []))

  // Get geolocation on mount
  useEffect(() => {
    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition(
        (pos) => {
          const p = { lat: pos.coords.latitude, lon: pos.coords.longitude }
          setUserPosition(p)
          setCenter(p)
        },
        () => {},
        { enableHighAccuracy: false, timeout: 5000 }
      )
    }
  }, [])

  // Parse URL params and auto-search on mount
  useEffect(() => {
    const urlQuery = parseUrlParams()
    if (urlQuery) {
      setQuery(urlQuery.q)
      setCenter({ lat: urlQuery.lat, lon: urlQuery.lon })
      doSearch(urlQuery.q, urlQuery.lat, urlQuery.lon, urlQuery.radius)
    }
  }, [])

  const doSearch = useCallback(async (q, lat, lon, radius) => {
    setLoading(true)
    setError(null)
    try {
      const data = await search(q, lat, lon, { radius })
      setResults(data.pois || [])
      setTotal(data.total || 0)
      setCached(data.cached || false)
      if (data.center) setCenter(data.center)

      setHistory(prev => {
        const updated = addHistory(prev, { query: q, lat, lon, timestamp: Date.now() })
        saveToStorage('history', updated)
        return updated
      })
    } catch (err) {
      setError(err.message)
      setResults([])
    } finally {
      setLoading(false)
    }
  }, [])

  const handleSearch = useCallback((q) => {
    doSearch(q, center.lat, center.lon)
  }, [center, doSearch])

  const handleFilter = useCallback((q) => {
    setQuery(q)
    doSearch(q, center.lat, center.lon)
  }, [center, doSearch])

  const toggleFavorite = useCallback((id) => {
    setFavorites(prev => {
      const updated = prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id]
      saveToStorage('favorites', updated)
      return updated
    })
  }, [])

  const selectPOI = useCallback((poi) => {
    setSelectedPOI(poi)
    setCenter({ lat: poi.lat, lon: poi.lon })
  }, [])

  return (
    <div className="app">
      <Header />
      <SearchBar query={query} onQueryChange={setQuery} onSearch={handleSearch} loading={loading} />
      <QuickFilters onFilter={handleFilter} activeQuery={query} />

      {history.length > 0 && !query && !results.length && (
        <div className="history-bar">
          <span className="history-label">История:</span>
          {history.slice(0, 5).map(h => (
            <button key={h.query} className="history-chip" onClick={() => { setQuery(h.query); handleSearch(h.query) }}>
              {h.query}
            </button>
          ))}
        </div>
      )}

      <div className="main-layout">
        <div className="panel-left">
          <ResultsList
            results={results}
            total={total}
            cached={cached}
            loading={loading}
            error={error}
            favorites={favorites}
            onToggleFavorite={toggleFavorite}
            onSelectPOI={selectPOI}
          />
          {!results.length && !loading && !error && (
            <div className="empty-state">
              <div className="empty-icon">📍</div>
              <div className="empty-text">Введите запрос, чтобы найти место</div>
              <div className="empty-hint">Например: «тихое кафе с террасой рядом с метро»</div>
            </div>
          )}
        </div>
        <div className="panel-right">
          <MapView
            results={results}
            center={center}
            userPosition={userPosition}
            onSelectPOI={selectPOI}
            selectedPOI={selectedPOI}
          />
        </div>
      </div>

      <DetailModal poi={selectedPOI} userPosition={userPosition} onClose={() => setSelectedPOI(null)} />
    </div>
  )
}
