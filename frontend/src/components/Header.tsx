import { useAuth } from '../contexts/AuthContext'
import Logo from './Logo'

interface HeaderProps {
  user: any
  onRefresh: () => void
  onAddProperty?: () => void
}

export default function Header({ user, onRefresh, onAddProperty }: HeaderProps) {
  const { logout } = useAuth()

  return (
    <header className="bg-white dark:bg-gray-800 shadow-sm border-b border-gray-200 dark:border-gray-700">
      <div className="container mx-auto px-4 py-3 flex items-center justify-between">
        <Logo size="sm" />
        <div className="flex items-center gap-4">
          {onAddProperty && (
            <button
              onClick={onAddProperty}
              className="px-4 py-2 text-sm bg-green-600 text-white rounded-md hover:bg-green-700"
            >
              + Add Property
            </button>
          )}
          <button
            onClick={onRefresh}
            className="px-4 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700"
          >
            Refresh
          </button>
          <div className="flex items-center gap-2">
            <span className="text-sm text-gray-600 dark:text-gray-300">{user?.username}</span>
            <button
              onClick={logout}
              className="px-4 py-2 text-sm bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-md hover:bg-gray-300 dark:hover:bg-gray-600"
            >
              Logout
            </button>
          </div>
        </div>
      </div>
    </header>
  )
}
