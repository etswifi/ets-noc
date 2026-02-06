import { useState, useEffect } from 'react'
import { apiClient } from '../api/client'
import DeviceDetailModal from './DeviceDetailModal'

interface DevicesListProps {
  devices: any[]
  propertyId: number
  onUpdate: () => void
}

export default function DevicesList({ devices, propertyId, onUpdate }: DevicesListProps) {
  const [showAddModal, setShowAddModal] = useState(false)
  const [editingDevice, setEditingDevice] = useState<any>(null)
  const [selectedDevice, setSelectedDevice] = useState<any>(null)
  const [deviceStatuses, setDeviceStatuses] = useState<Record<number, any>>({})
  const [formData, setFormData] = useState({
    name: '',
    hostname: '',
    device_type: 'wap',
    is_critical: false,
    check_interval: 60,
    retries: 3,
    timeout: 10000,
  })

  useEffect(() => {
    loadDeviceStatuses()
    const interval = setInterval(loadDeviceStatuses, 10000) // Refresh every 10s
    return () => clearInterval(interval)
  }, [devices])

  const loadDeviceStatuses = async () => {
    const statuses: Record<number, any> = {}
    await Promise.all(
      devices.map(async (device) => {
        try {
          const status = await apiClient.getDeviceStatus(device.id)
          statuses[device.id] = status
        } catch (error) {
          console.error(`Failed to load status for device ${device.id}:`, error)
        }
      })
    )
    setDeviceStatuses(statuses)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      if (editingDevice) {
        await apiClient.updateDevice(editingDevice.id, {
          ...formData,
          property_id: propertyId,
        })
      } else {
        await apiClient.createDevice({
          ...formData,
          property_id: propertyId,
        })
      }
      setShowAddModal(false)
      setEditingDevice(null)
      setFormData({ name: '', hostname: '', device_type: 'wap', is_critical: false, check_interval: 60, retries: 3, timeout: 10000 })
      onUpdate()
    } catch (error: any) {
      alert(error.message)
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this device?')) return
    try {
      await apiClient.deleteDevice(id)
      onUpdate()
    } catch (error: any) {
      alert(error.message)
    }
  }

  const openEditModal = (device: any) => {
    setEditingDevice(device)
    setFormData({
      name: device.name,
      hostname: device.hostname,
      device_type: device.device_type,
      is_critical: device.is_critical,
      check_interval: device.check_interval || 60,
      retries: device.retries || 3,
      timeout: device.timeout || 10000,
    })
    setShowAddModal(true)
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h3 className="text-lg font-semibold">Devices ({devices.length})</h3>
        <button
          onClick={() => setShowAddModal(true)}
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
        >
          Add Device
        </button>
      </div>

      <div className="space-y-2">
        {devices
          .sort((a, b) => {
            // Sort by IP address (convert to number for proper sorting)
            const ipToNum = (ip: string) => {
              const parts = ip.split('.').map(Number)
              return (parts[0] || 0) * 16777216 + (parts[1] || 0) * 65536 + (parts[2] || 0) * 256 + (parts[3] || 0)
            }
            return ipToNum(a.hostname) - ipToNum(b.hostname)
          })
          .map((device) => {
          const status = deviceStatuses[device.id]
          const isOnline = status?.status === 'online'
          return (
            <div
              key={device.id}
              className="bg-gray-50 rounded-lg p-4 flex items-center justify-between hover:bg-gray-100 cursor-pointer transition-colors"
              onClick={() => setSelectedDevice(device)}
            >
              <div className="flex items-center gap-3 flex-1">
                {/* Status Indicator */}
                <div className="flex flex-col items-center">
                  <div
                    className={`w-4 h-4 rounded-full ${
                      isOnline ? 'bg-green-500' : 'bg-red-500'
                    } animate-pulse`}
                  />
                  {status?.response_time && (
                    <div className="text-xs text-gray-500 mt-1">
                      {status.response_time}ms
                    </div>
                  )}
                </div>

                {/* Device Info */}
                <div className="flex-1">
                  <div className="flex items-center gap-2">
                    <h4 className="font-medium">{device.name}</h4>
                    {device.is_critical && (
                      <span className="text-xs bg-red-600 text-white px-2 py-1 rounded-full">
                        CRITICAL
                      </span>
                    )}
                    <span
                      className={`text-xs px-2 py-1 rounded-full ${
                        isOnline
                          ? 'bg-green-100 text-green-700'
                          : 'bg-red-100 text-red-700'
                      }`}
                    >
                      {isOnline ? 'ONLINE' : 'OFFLINE'}
                    </span>
                  </div>
                  <div className="text-sm text-gray-600 mt-1">
                    <a
                      href={`https://${device.hostname}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-blue-600 hover:underline"
                      onClick={(e) => e.stopPropagation()}
                    >
                      {device.hostname}
                    </a>
                    {' • '}{device.device_type}
                    {device.tags && device.tags.length > 0 && (
                      <span className="ml-2">
                        {device.tags.map((tag: string) => (
                          <span
                            key={tag}
                            className="inline-block bg-gray-200 px-2 py-0.5 rounded text-xs ml-1"
                          >
                            {tag}
                          </span>
                        ))}
                      </span>
                    )}
                  </div>
                </div>
              </div>

              {/* Actions */}
              <div className="flex gap-2" onClick={(e) => e.stopPropagation()}>
                <button
                  onClick={() => openEditModal(device)}
                  className="px-3 py-1 text-sm bg-gray-200 text-gray-700 rounded hover:bg-gray-300"
                >
                  Edit
                </button>
                <button
                  onClick={() => handleDelete(device.id)}
                  className="px-3 py-1 text-sm bg-red-600 text-white rounded hover:bg-red-700"
                >
                  Delete
                </button>
              </div>
            </div>
          )
        })}
      </div>

      {devices.length === 0 && (
        <div className="text-center py-12 text-gray-500">
          No devices configured for this property
        </div>
      )}

      {showAddModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full">
            <h3 className="text-xl font-bold mb-4">
              {editingDevice ? 'Edit Device' : 'Add Device'}
            </h3>
            <form onSubmit={handleSubmit}>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Device Name
                  </label>
                  <input
                    type="text"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Hostname / IP
                  </label>
                  <input
                    type="text"
                    value={formData.hostname}
                    onChange={(e) => setFormData({ ...formData, hostname: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Device Type
                  </label>
                  <select
                    value={formData.device_type}
                    onChange={(e) => setFormData({ ...formData, device_type: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                  >
                    <option value="wap">WAP</option>
                    <option value="switch">Switch</option>
                    <option value="router">Router</option>
                  </select>
                </div>
                <div className="flex items-center">
                  <input
                    type="checkbox"
                    checked={formData.is_critical}
                    onChange={(e) => setFormData({ ...formData, is_critical: e.target.checked })}
                    className="mr-2"
                  />
                  <label className="text-sm text-gray-700">Mark as Critical Device</label>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Check Interval (seconds)
                  </label>
                  <input
                    type="number"
                    value={formData.check_interval}
                    onChange={(e) => setFormData({ ...formData, check_interval: parseInt(e.target.value) || 60 })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                    min="10"
                    required
                  />
                  <p className="text-xs text-gray-500 mt-1">How often to check device status (default: 60)</p>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Retries
                  </label>
                  <input
                    type="number"
                    value={formData.retries}
                    onChange={(e) => setFormData({ ...formData, retries: parseInt(e.target.value) || 3 })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                    min="1"
                    max="10"
                    required
                  />
                  <p className="text-xs text-gray-500 mt-1">Number of retry attempts (default: 3)</p>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Timeout (milliseconds)
                  </label>
                  <input
                    type="number"
                    value={formData.timeout}
                    onChange={(e) => setFormData({ ...formData, timeout: parseInt(e.target.value) || 10000 })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                    min="1000"
                    max="60000"
                    required
                  />
                  <p className="text-xs text-gray-500 mt-1">Ping timeout in ms (default: 10000)</p>
                </div>
              </div>
              <div className="flex justify-end gap-2 mt-6">
                <button
                  type="button"
                  onClick={() => {
                    setShowAddModal(false)
                    setEditingDevice(null)
                    setFormData({ name: '', hostname: '', device_type: 'wap', is_critical: false, check_interval: 60, retries: 3, timeout: 10000 })
                  }}
                  className="px-4 py-2 bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
                >
                  {editingDevice ? 'Update' : 'Add'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {selectedDevice && (
        <DeviceDetailModal
          device={selectedDevice}
          onClose={() => setSelectedDevice(null)}
        />
      )}
    </div>
  )
}
