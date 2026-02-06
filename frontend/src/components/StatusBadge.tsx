interface StatusBadgeProps {
  status: string
}

export default function StatusBadge({ status }: StatusBadgeProps) {
  const statusStyles = {
    red: 'bg-red-600 text-white',
    yellow: 'bg-yellow-500 text-white',
    green: 'bg-green-600 text-white',
  }

  return (
    <span
      className={`px-2 py-1 rounded-full text-xs font-medium uppercase ${
        statusStyles[status as keyof typeof statusStyles] || 'bg-gray-500 text-white'
      }`}
    >
      {status}
    </span>
  )
}
