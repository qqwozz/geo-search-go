import { useEffect, useRef } from 'react'
import { MapContainer, TileLayer, Marker, Popup, useMap } from 'react-leaflet'
import L from 'leaflet'
import './MapView.css'

function createIcon(color) {
  return L.divIcon({
    className: 'custom-marker',
    html: `<div style="width:14px;height:14px;background:${color};border:2px solid #fff;border-radius:50%;box-shadow:0 1px 4px rgba(0,0,0,0.3)"></div>`,
    iconSize: [14, 14],
    iconAnchor: [7, 7],
  })
}

const userIcon = createIcon('#4a90d9')
const poiIcon = createIcon('#e74c3c')
const selectedIcon = createIcon('#ff9800')

function FlyTo({ center }) {
  const map = useMap()
  useEffect(() => {
    if (center) map.flyTo([center.lat, center.lon], map.getZoom(), { duration: 0.5 })
  }, [center, map])
  return null
}

export default function MapView({ results, center, userPosition, onSelectPOI, selectedPOI }) {
  const defaultCenter = center || { lat: 55.755, lon: 37.615 }

  return (
    <div className="map-container">
      <MapContainer
        center={[defaultCenter.lat, defaultCenter.lon]}
        zoom={14}
        className="leaflet-map"
        zoomControl={true}
      >
        <TileLayer
          attribution='&copy; <a href="https://osm.org/copyright">OpenStreetMap</a>'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        <FlyTo center={center} />

        {userPosition && (
          <Marker position={[userPosition.lat, userPosition.lon]} icon={userIcon}>
            <Popup>Вы здесь</Popup>
          </Marker>
        )}

        {results.map(poi => (
          <Marker
            key={poi.id}
            position={[poi.lat, poi.lon]}
            icon={selectedPOI?.id === poi.id ? selectedIcon : poiIcon}
            eventHandlers={{ click: () => onSelectPOI(poi) }}
          >
            <Popup>
              <strong>{poi.name}</strong><br />
              {poi.rating} ★ · {poi.address || ''}
            </Popup>
          </Marker>
        ))}
      </MapContainer>
    </div>
  )
}
