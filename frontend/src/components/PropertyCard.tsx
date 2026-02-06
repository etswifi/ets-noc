import StatusBadge from './StatusBadge'

interface PropertyCardProps {
  property: any
  onClick: () => void
}

export default function PropertyCard({ property, onClick }: PropertyCardProps) {
  const statusColors = {
    red: 'border-red-500 bg-red-50',
    yellow: 'border-yellow-500 bg-yellow-50',
    green: 'border-green-500 bg-green-50',
  }

  return (
    <div
      onClick={onClick}
      className={`bg-white rounded-lg shadow-md p-4 cursor-pointer hover:shadow-lg transition-shadow border-l-4 ${
        statusColors[property.status as keyof typeof statusColors] || 'border-gray-500'
      }`}
    >
      <div className="flex items-start justify-between mb-2">
        <h3 className="text-lg font-semibold text-gray-900 truncate flex-1">
          {property.name}
        </h3>
        <StatusBadge status={property.status} />
      </div>

      {property.address && (
        <p className="text-sm text-gray-600 mb-3 line-clamp-2">{property.address}</p>
      )}

      <div className="flex items-center justify-between text-sm">
        <div>
          <span className="text-green-600 font-medium">{property.online_count || 0}</span>
          <span className="text-gray-500"> / </span>
          <span className="text-gray-700">{property.total_count || 0}</span>
          <span className="text-gray-500"> devices</span>
        </div>
        {property.critical_offline && (
          <span className="text-xs bg-red-600 text-white px-2 py-1 rounded-full">
            CRITICAL
          </span>
        )}
      </div>

      {property.offline_count > 0 && (
        <div className="mt-2 text-xs text-red-600">
          {property.offline_count} device{property.offline_count > 1 ? 's' : ''} offline
        </div>
      )}
    </div>
  )
}
