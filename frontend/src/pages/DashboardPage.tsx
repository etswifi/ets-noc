import { useState, useEffect } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { apiClient } from '../api/client'
import PropertyCard from '../components/PropertyCard'
import PropertyDetailModal from '../components/PropertyDetailModal'
import PropertyModal from '../components/PropertyModal'
import Header from '../components/Header'

export default function DashboardPage() {
  const { user } = useAuth()
  const [dashboard, setDashboard] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [selectedProperty, setSelectedProperty] = useState<any>(null)
  const [showPropertyModal, setShowPropertyModal] = useState(false)
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [searchQuery, setSearchQuery] = useState('')

  useEffect(() => {
    loadDashboard()
    const interval = setInterval(loadDashboard, 30000) // Refresh every 30 seconds
    return () => clearInterval(interval)
  }, [])

  const loadDashboard = async () => {
    try {
      const data = await apiClient.getDashboard()
      setDashboard(data)
    } catch (error) {
      console.error('Failed to load dashboard:', error)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
        <div className="text-xl text-gray-900 dark:text-white">Loading dashboard...</div>
      </div>
    )
  }

  if (!dashboard) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
        <div className="text-xl text-red-600 dark:text-red-400">Failed to load dashboard</div>
      </div>
    )
  }

  const filteredProperties = dashboard.properties.filter((property: any) => {
    const matchesStatus = statusFilter === 'all' || property.status === statusFilter
    const matchesSearch = property.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      property.address?.toLowerCase().includes(searchQuery.toLowerCase())
    return matchesStatus && matchesSearch
  })

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <Header user={user} onRefresh={loadDashboard} onAddProperty={() => setShowPropertyModal(true)} />

      <div className="container mx-auto px-4 py-6">
        {/* Summary Cards */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
            <div className="text-2xl font-bold text-gray-900 dark:text-white">{dashboard.summary.total_properties}</div>
            <div className="text-gray-600 dark:text-gray-400">Total Properties</div>
          </div>
          <div className="bg-red-100 dark:bg-red-900/30 rounded-lg shadow p-4">
            <div className="text-2xl font-bold text-red-600 dark:text-red-400">{dashboard.summary.red_count}</div>
            <div className="text-gray-600 dark:text-gray-400">Critical</div>
          </div>
          <div className="bg-yellow-100 dark:bg-yellow-900/30 rounded-lg shadow p-4">
            <div className="text-2xl font-bold text-yellow-600 dark:text-yellow-400">{dashboard.summary.yellow_count}</div>
            <div className="text-gray-600 dark:text-gray-400">Warning</div>
          </div>
          <div className="bg-green-100 dark:bg-green-900/30 rounded-lg shadow p-4">
            <div className="text-2xl font-bold text-green-600 dark:text-green-400">{dashboard.summary.green_count}</div>
            <div className="text-gray-600 dark:text-gray-400">Healthy</div>
          </div>
        </div>

        {/* Filters */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-6">
          <div className="flex flex-col md:flex-row gap-4">
            <input
              type="text"
              placeholder="Search properties..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="flex-1 px-4 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <div className="flex gap-2">
              <button
                onClick={() => setStatusFilter('all')}
                className={`px-4 py-2 rounded-md ${
                  statusFilter === 'all'
                    ? 'bg-blue-600 text-white'
                    : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600'
                }`}
              >
                All
              </button>
              <button
                onClick={() => setStatusFilter('red')}
                className={`px-4 py-2 rounded-md ${
                  statusFilter === 'red'
                    ? 'bg-red-600 text-white'
                    : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600'
                }`}
              >
                Critical
              </button>
              <button
                onClick={() => setStatusFilter('yellow')}
                className={`px-4 py-2 rounded-md ${
                  statusFilter === 'yellow'
                    ? 'bg-yellow-600 text-white'
                    : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600'
                }`}
              >
                Warning
              </button>
              <button
                onClick={() => setStatusFilter('green')}
                className={`px-4 py-2 rounded-md ${
                  statusFilter === 'green'
                    ? 'bg-green-600 text-white'
                    : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600'
                }`}
              >
                Healthy
              </button>
            </div>
          </div>
        </div>

        {/* Property Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {filteredProperties.map((property: any) => (
            <PropertyCard
              key={property.id}
              property={property}
              onClick={() => setSelectedProperty(property)}
            />
          ))}
        </div>

        {filteredProperties.length === 0 && (
          <div className="text-center py-12 text-gray-500 dark:text-gray-400">
            No properties match your filters
          </div>
        )}
      </div>

      {selectedProperty && (
        <PropertyDetailModal
          property={selectedProperty}
          onClose={() => setSelectedProperty(null)}
          onUpdate={loadDashboard}
        />
      )}

      {showPropertyModal && (
        <PropertyModal
          onClose={() => setShowPropertyModal(false)}
          onSuccess={loadDashboard}
        />
      )}
    </div>
  )
}
