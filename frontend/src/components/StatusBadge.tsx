interface StatusBadgeProps {
  status: string
}

export default function StatusBadge({ status }: StatusBadgeProps) {
  const statusConfig = {
    red: { label: 'Critical', color: 'bg-red-600 text-white' },
    yellow: { label: 'Warning', color: 'bg-yellow-500 text-white' },
    green: { label: 'Healthy', color: 'bg-green-600 text-white' },
  }

  const config = statusConfig[status as keyof typeof statusConfig] || {
    label: status,
    color: 'bg-gray-500 text-white',
  }

  return (
    <span
      className={`px-2 py-1 rounded-full text-xs font-medium uppercase ${config.color}`}
    >
      {config.label}
    </span>
  )
}
