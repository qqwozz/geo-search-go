import { useState, useEffect, useRef } from 'react'
import { autocomplete } from '../api'
import './SearchBar.css'

export default function SearchBar({ query, onQueryChange, onSearch, loading }) {
  const [suggestions, setSuggestions] = useState([])
  const [showDropdown, setShowDropdown] = useState(false)
  const [highlightIdx, setHighlightIdx] = useState(-1)
  const wrapperRef = useRef(null)
  const debounceRef = useRef(null)

  useEffect(() => {
    if (!query) {
      setSuggestions([])
      return
    }
    clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => {
      autocomplete(query).then(r => {
        setSuggestions(r.suggestions || [])
        setShowDropdown(true)
      }).catch(() => setSuggestions([]))
    }, 300)
    return () => clearTimeout(debounceRef.current)
  }, [query])

  useEffect(() => {
    const handler = (e) => {
      if (wrapperRef.current && !wrapperRef.current.contains(e.target)) {
        setShowDropdown(false)
      }
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  const handleSubmit = (e) => {
    e.preventDefault()
    setShowDropdown(false)
    if (query.trim()) onSearch(query.trim())
  }

  const handleKeyDown = (e) => {
    if (!showDropdown || !suggestions.length) return
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      setHighlightIdx(i => (i + 1) % suggestions.length)
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setHighlightIdx(i => (i - 1 + suggestions.length) % suggestions.length)
    } else if (e.key === 'Enter' && highlightIdx >= 0) {
      e.preventDefault()
      onQueryChange(suggestions[highlightIdx])
      setShowDropdown(false)
      onSearch(suggestions[highlightIdx])
    } else if (e.key === 'Escape') {
      setShowDropdown(false)
    }
  }

  const pickSuggestion = (s) => {
    onQueryChange(s)
    setShowDropdown(false)
    onSearch(s)
  }

  return (
    <div className="searchbar-wrapper" ref={wrapperRef}>
      <form className="searchbar" onSubmit={handleSubmit}>
        <input
          className="searchbar-input"
          type="text"
          placeholder="Найди место: «тихое кафе с террасой»"
          value={query}
          onChange={(e) => { onQueryChange(e.target.value); setHighlightIdx(-1) }}
          onKeyDown={handleKeyDown}
          onFocus={() => suggestions.length && setShowDropdown(true)}
          maxLength={500}
        />
        <button className="searchbar-btn" type="submit" disabled={loading || !query.trim()}>
          {loading ? '...' : 'Найти'}
        </button>
      </form>
      {showDropdown && suggestions.length > 0 && (
        <ul className="suggestions">
          {suggestions.map((s, i) => (
            <li
              key={s}
              className={`suggestion ${i === highlightIdx ? 'active' : ''}`}
              onMouseDown={() => pickSuggestion(s)}
              onMouseEnter={() => setHighlightIdx(i)}
            >
              {s}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
