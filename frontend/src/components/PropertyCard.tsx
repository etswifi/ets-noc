import StatusBadge from './StatusBadge'

interface PropertyCardProps {
  property: any
  onClick: () => void
}

export default function PropertyCard({ property, onClick }: PropertyCardProps) {
  const statusColors = {
    red: 'border-red-500 bg-red-50 dark:bg-red-900/20',
    yellow: 'border-yellow-500 bg-yellow-50 dark:bg-yellow-900/20',
    green: 'border-green-500 bg-green-50 dark:bg-green-900/20',
  }

  return (
    <div
      onClick={onClick}
      className={`bg-white dark:bg-gray-800 rounded-lg shadow-md p-4 cursor-pointer hover:shadow-lg transition-shadow border-l-4 ${
        statusColors[property.status as keyof typeof statusColors] || 'border-gray-500'
      }`}
    >
      <div className="flex items-start justify-between mb-2">
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white truncate flex-1">
          {property.name}
        </h3>
        <StatusBadge status={property.status} />
      </div>

      {property.address && (
        <a
          href={`https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(property.address)}`}
          target="_blank"
          rel="noopener noreferrer"
          onClick={(e) => e.stopPropagation()}
          className="text-sm text-blue-600 dark:text-blue-400 hover:underline mb-3 line-clamp-2 block"
        >
          {property.address}
        </a>
      )}

      <div className="flex items-center justify-between text-sm">
        <div>
          <span className="text-green-600 dark:text-green-400 font-medium">{property.online_count || 0}</span>
          <span className="text-gray-500 dark:text-gray-400"> / </span>
          <span className="text-gray-700 dark:text-gray-300">{property.total_count || 0}</span>
          <span className="text-gray-500 dark:text-gray-400"> devices</span>
        </div>
        {property.critical_offline && (
          <span className="text-xs bg-red-600 text-white px-2 py-1 rounded-full">
            CRITICAL
          </span>
        )}
      </div>

      {property.offline_count > 0 && (
        <div className="mt-2 text-xs text-red-600 dark:text-red-400">
          {property.offline_count} device{property.offline_count > 1 ? 's' : ''} offline
        </div>
      )}
    </div>
  )
}
