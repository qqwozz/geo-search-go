import './DetailModal.css'

const FEATURE_LABELS = {
  wifi: 'Wi-Fi', outlets: 'Розетки', terrace: 'Терраса', parking: 'Парковка',
  live_music: 'Музыка', breakfast: 'Завтраки', quiet: 'Тихо',
  family_friendly: 'Семейное', romantic: 'Романтично', dog_friendly: 'С собакой',
}

const NOISE_LABELS = { quiet: 'Тихо', medium: 'Средне', loud: 'Шумно' }
const PRICE_LABELS = { 1: 'Бюджетно', 2: 'Средне', 3: 'Дорого', 4: 'Очень дорого' }

export default function DetailModal({ poi, userPosition, onClose }) {
  if (!poi) return null

  const directionsUrl = userPosition
    ? `https://yandex.ru/maps/?rtext=${userPosition.lat},${userPosition.lon}~${poi.lat},${poi.lon}&rtt=auto`
    : `https://yandex.ru/maps/?text=${encodeURIComponent(poi.address || poi.name)}`

  const shareUrl = `${window.location.origin}?q=${encodeURIComponent(poi.name)}&lat=${poi.lat}&lon=${poi.lon}`

  const handleShare = () => {
    navigator.clipboard.writeText(shareUrl).then(
      () => alert('Ссылка скопирована!'),
      () => prompt('Скопируйте ссылку:', shareUrl)
    )
  }

  const activeFeatures = Object.entries(poi.features || {})
    .filter(([, v]) => v)
    .map(([k]) => k)

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={e => e.stopPropagation()}>
        <button className="modal-close" onClick={onClose}>✕</button>

        <h2 className="modal-title">{poi.name}</h2>

        {poi.name_en && <div className="modal-name-en">{poi.name_en}</div>}

        <div className="modal-meta">
          <span className="modal-rating">★ {poi.rating}</span>
          <span className="modal-reviews">{poi.review_count} отзывов</span>
          {poi.price_level && <span className="modal-price">{PRICE_LABELS[poi.price_level] || `Уровень ${poi.price_level}`}</span>}
          {poi.noise_level && <span className="modal-noise">{NOISE_LABELS[poi.noise_level] || poi.noise_level}</span>}
        </div>

        {poi.address && <div className="modal-detail"><strong>Адрес:</strong> {poi.address}</div>}
        {poi.phone && <div className="modal-detail"><strong>Телефон:</strong> {poi.phone}</div>}
        {poi.website && (
          <div className="modal-detail">
            <strong>Сайт:</strong> <a href={poi.website} target="_blank" rel="noreferrer">{poi.website}</a>
          </div>
        )}
        {poi.opening_hours && <div className="modal-detail"><strong>Часы:</strong> {poi.opening_hours}</div>}

        {activeFeatures.length > 0 && (
          <div className="modal-features">
            {activeFeatures.map(f => (
              <span key={f} className="modal-feature-tag">{FEATURE_LABELS[f] || f}</span>
            ))}
          </div>
        )}

        {poi.explanation && (
          <div className="modal-explanation">{poi.explanation}</div>
        )}

        <div className="modal-actions">
          <a className="modal-btn primary" href={directionsUrl} target="_blank" rel="noreferrer">
            Построить маршрут
          </a>
          <button className="modal-btn secondary" onClick={handleShare}>
            Поделиться
          </button>
        </div>
      </div>
    </div>
  )
}
