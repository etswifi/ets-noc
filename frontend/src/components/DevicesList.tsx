import { useState } from 'react'
import { apiClient } from '../api/client'

interface DevicesListProps {
  devices: any[]
  propertyId: number
  onUpdate: () => void
}

export default function DevicesList({ devices, propertyId, onUpdate }: DevicesListProps) {
  const [showAddModal, setShowAddModal] = useState(false)
  const [editingDevice, setEditingDevice] = useState<any>(null)
  const [formData, setFormData] = useState({
    name: '',
    hostname: '',
    device_type: 'wap',
    is_critical: false,
  })

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
      setFormData({ name: '', hostname: '', device_type: 'wap', is_critical: false })
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
        {devices.map((device) => (
          <div key={device.id} className="bg-gray-50 rounded-lg p-4 flex items-center justify-between">
            <div className="flex-1">
              <div className="flex items-center gap-2">
                <h4 className="font-medium">{device.name}</h4>
                {device.is_critical && (
                  <span className="text-xs bg-red-600 text-white px-2 py-1 rounded-full">
                    CRITICAL
                  </span>
                )}
              </div>
              <div className="text-sm text-gray-600 mt-1">
                {device.hostname} • {device.device_type}
              </div>
            </div>
            <div className="flex gap-2">
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
        ))}
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
              </div>
              <div className="flex justify-end gap-2 mt-6">
                <button
                  type="button"
                  onClick={() => {
                    setShowAddModal(false)
                    setEditingDevice(null)
                    setFormData({ name: '', hostname: '', device_type: 'wap', is_critical: false })
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
    </div>
  )
}
