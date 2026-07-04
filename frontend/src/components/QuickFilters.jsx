import './QuickFilters.css'

const FILTERS = [
  { label: 'Кофе', query: 'кафе' },
  { label: 'Завтрак', query: 'завтрак' },
  { label: 'Работа', query: 'кафе для работы' },
  { label: 'Терраса', query: 'кафе с террасой' },
  { label: 'Ужин', query: 'ресторан ужин' },
]

export default function QuickFilters({ onFilter, activeQuery }) {
  return (
    <div className="quick-filters">
      {FILTERS.map(f => (
        <button
          key={f.query}
          className={`chip ${activeQuery === f.query ? 'active' : ''}`}
          onClick={() => onFilter(f.query)}
        >
          {f.label}
        </button>
      ))}
    </div>
  )
}
