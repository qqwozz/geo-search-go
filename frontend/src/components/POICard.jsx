import './POICard.css'

const CATEGORY_LABELS = {
  cafe: 'Кафе',
  restaurant: 'Ресторан',
  bar: 'Бар',
  fast_food: 'Фастфуд',
  park: 'Парк',
}

const FEATURE_LABELS = {
  wifi: 'Wi-Fi',
  outlets: 'Розетки',
  terrace: 'Терраса',
  parking: 'Парковка',
  live_music: 'Музыка',
  breakfast: 'Завтраки',
  quiet: 'Тихо',
  family_friendly: 'Семейное',
  romantic: 'Романтично',
  dog_friendly: 'С собакой',
}

function formatDistance(m) {
  return m < 1000 ? `${Math.round(m)} м` : `${(m / 1000).toFixed(1)} км`
}

function renderStars(rating) {
  const full = Math.floor(rating)
  const half = rating - full >= 0.5
  return '★'.repeat(full) + (half ? '½' : '') + '☆'.repeat(5 - full - (half ? 1 : 0))
}

export default function POICard({ poi, isFavorite, onToggleFavorite, onSelect }) {
  const activeFeatures = Object.entries(poi.features || {})
    .filter(([, v]) => v)
    .map(([k]) => k)

  return (
    <div className="poi-card" onClick={() => onSelect(poi)}>
      <div className="poi-card-header">
        <div className="poi-card-title">
          <span className="poi-name">{poi.name}</span>
          <span className="poi-category">{CATEGORY_LABELS[poi.category] || poi.category}</span>
        </div>
        <button
          className={`fav-btn ${isFavorite ? 'active' : ''}`}
          onClick={(e) => { e.stopPropagation(); onToggleFavorite(poi.id) }}
          title={isFavorite ? 'Убрать из избранного' : 'В избранное'}
        >
          {isFavorite ? '♥' : '♡'}
        </button>
      </div>

      <div className="poi-card-rating">
        <span className="stars">{renderStars(poi.rating)}</span>
        <span className="rating-num">{poi.rating}</span>
        <span className="review-count">({poi.review_count})</span>
        {poi.distance_meters != null && (
          <span className="distance">{formatDistance(poi.distance_meters)}</span>
        )}
      </div>

      {poi.address && <div className="poi-address">{poi.address}</div>}

      <div className="poi-tags">
        {activeFeatures.map(f => (
          <span key={f} className="tag">{FEATURE_LABELS[f] || f}</span>
        ))}
      </div>

      {poi.explanation && (
        <div className="poi-explanation">{poi.explanation}</div>
      )}
    </div>
  )
}
