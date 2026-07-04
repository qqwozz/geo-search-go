import POICard from './POICard'
import './ResultsList.css'

export default function ResultsList({ results, total, cached, loading, error, favorites, onToggleFavorite, onSelectPOI }) {
  if (loading) return <div className="results-status">Поиск мест...</div>
  if (error) return <div className="results-status error">Ошибка: {error}</div>
  if (!results.length) return null

  return (
    <div className="results-list">
      <div className="results-header">
        <span>{total} мест{total === 1 ? 'о' : total < 5 ? 'а' : ''}</span>
        {cached && <span className="cached-badge">из кеша</span>}
      </div>
      {results.map(poi => (
        <POICard
          key={poi.id}
          poi={poi}
          isFavorite={favorites.includes(poi.id)}
          onToggleFavorite={onToggleFavorite}
          onSelect={onSelectPOI}
        />
      ))}
    </div>
  )
}
