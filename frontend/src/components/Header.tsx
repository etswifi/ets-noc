import { useAuth } from '../contexts/AuthContext'

interface HeaderProps {
  user: any
  onRefresh: () => void
}

export default function Header({ user, onRefresh }: HeaderProps) {
  const { logout } = useAuth()

  return (
    <header className="bg-white shadow-sm">
      <div className="container mx-auto px-4 py-4 flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">ETS Properties Monitoring</h1>
        <div className="flex items-center gap-4">
          <button
            onClick={onRefresh}
            className="px-4 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700"
          >
            Refresh
          </button>
          <div className="flex items-center gap-2">
            <span className="text-sm text-gray-600">{user?.username}</span>
            <button
              onClick={logout}
              className="px-4 py-2 text-sm bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300"
            >
              Logout
            </button>
          </div>
        </div>
      </div>
    </header>
  )
}
