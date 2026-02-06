import { useState, useEffect } from 'react'
import { apiClient } from '../api/client'
import DevicesList from './DevicesList'
import ContactsList from './ContactsList'
import AttachmentsList from './AttachmentsList'
import AddressAutocomplete from './AddressAutocomplete'

interface PropertyDetailModalProps {
  property: any
  onClose: () => void
  onUpdate: () => void
}

export default function PropertyDetailModal({ property, onClose, onUpdate }: PropertyDetailModalProps) {
  const [activeTab, setActiveTab] = useState('details')
  const [devices, setDevices] = useState<any[]>([])
  const [contacts, setContacts] = useState<any[]>([])
  const [attachments, setAttachments] = useState<any[]>([])
  const [loading, setLoading] = useState(false)
  const [syncing, setSyncing] = useState(false)
  const [editing, setEditing] = useState(false)
  const [formData, setFormData] = useState({
    name: property.name || '',
    address: property.address || '',
    notes: property.notes || '',
    isp_company_name: property.isp_company_name || '',
    isp_account_info: property.isp_account_info || '',
  })

  useEffect(() => {
    if (activeTab === 'devices') {
      loadDevices()
    } else if (activeTab === 'contacts') {
      loadContacts()
    } else if (activeTab === 'attachments') {
      loadAttachments()
    }
  }, [activeTab])

  const loadDevices = async () => {
    setLoading(true)
    try {
      const data = await apiClient.getPropertyDevices(property.id)
      setDevices(data)
    } catch (error) {
      console.error('Failed to load devices:', error)
    } finally {
      setLoading(false)
    }
  }

  const loadContacts = async () => {
    setLoading(true)
    try {
      const data = await apiClient.getContacts(property.id)
      setContacts(data)
    } catch (error) {
      console.error('Failed to load contacts:', error)
    } finally {
      setLoading(false)
    }
  }

  const loadAttachments = async () => {
    setLoading(true)
    try {
      const data = await apiClient.getAttachments(property.id)
      setAttachments(data)
    } catch (error) {
      console.error('Failed to load attachments:', error)
    } finally {
      setLoading(false)
    }
  }

  const syncDevices = async () => {
    setSyncing(true)
    try {
      await apiClient.syncDevicesFromPfSense(property.id)
      await loadDevices()
      alert('Devices synced successfully from pfSense')
    } catch (error) {
      console.error('Failed to sync devices:', error)
      alert(`Failed to sync devices: ${error}`)
    } finally {
      setSyncing(false)
    }
  }

  const handleSave = async () => {
    try {
      await apiClient.updateProperty(property.id, formData)
      setEditing(false)
      onUpdate()
      alert('Property updated successfully')
    } catch (error) {
      console.error('Failed to update property:', error)
      alert(`Failed to update property: ${error}`)
    }
  }

  const handleCancel = () => {
    setFormData({
      name: property.name || '',
      address: property.address || '',
      notes: property.notes || '',
      isp_company_name: property.isp_company_name || '',
      isp_account_info: property.isp_account_info || '',
    })
    setEditing(false)
  }

  const tabs = [
    { id: 'details', label: 'Details' },
    { id: 'devices', label: 'Devices' },
    { id: 'contacts', label: 'Contacts' },
    { id: 'attachments', label: 'Attachments' },
  ]

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-6xl w-full max-h-[90vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
          <div>
            <h2 className="text-2xl font-bold text-gray-900 dark:text-white">{property.name}</h2>
            {property.address && (
              <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">{property.address}</p>
            )}
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">ID: {property.id} • Subnet: {property.subnet || 'N/A'}</p>
          </div>
          <div className="flex items-center gap-2">
            {activeTab === 'details' && !editing && (
              <button
                onClick={() => setEditing(true)}
                className="px-4 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700"
              >
                Edit
              </button>
            )}
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 text-2xl font-bold"
            >
              ×
            </button>
          </div>
        </div>

        {/* Tabs */}
        <div className="border-b border-gray-200 dark:border-gray-700">
          <div className="flex space-x-8 px-6">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`py-4 px-1 border-b-2 font-medium text-sm ${
                  activeTab === tab.id
                    ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                    : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-gray-600'
                }`}
              >
                {tab.label}
              </button>
            ))}
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6">
          {loading ? (
            <div className="text-center py-12">
              <div className="text-lg text-gray-600 dark:text-gray-400">Loading...</div>
            </div>
          ) : (
            <>
              {activeTab === 'details' && (
                <div className="space-y-4">
                  {editing ? (
                    <>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                          Property Name
                        </label>
                        <input
                          type="text"
                          value={formData.name}
                          onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                          className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                          Address
                        </label>
                        <AddressAutocomplete
                          value={formData.address}
                          onChange={(address) => setFormData({ ...formData, address })}
                          className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                          ISP Company
                        </label>
                        <input
                          type="text"
                          value={formData.isp_company_name}
                          onChange={(e) => setFormData({ ...formData, isp_company_name: e.target.value })}
                          className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                          ISP Account Info
                        </label>
                        <textarea
                          value={formData.isp_account_info}
                          onChange={(e) => setFormData({ ...formData, isp_account_info: e.target.value })}
                          className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                          rows={3}
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                          Notes
                        </label>
                        <textarea
                          value={formData.notes}
                          onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                          className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                          rows={4}
                        />
                      </div>
                      <div className="flex gap-2 pt-4">
                        <button
                          onClick={handleSave}
                          className="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700"
                        >
                          Save
                        </button>
                        <button
                          onClick={handleCancel}
                          className="px-4 py-2 bg-gray-300 dark:bg-gray-600 text-gray-700 dark:text-gray-300 rounded-md hover:bg-gray-400 dark:hover:bg-gray-500"
                        >
                          Cancel
                        </button>
                      </div>
                    </>
                  ) : (
                    <>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                          Property ID
                        </label>
                        <div className="text-gray-900 dark:text-white">{property.id}</div>
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                          Subnet
                        </label>
                        <div className="text-gray-900 dark:text-white">{property.subnet || 'N/A'}</div>
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                          Property Name
                        </label>
                        <div className="text-gray-900 dark:text-white">{property.name}</div>
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                          Address
                        </label>
                        {property.address ? (
                          <a
                            href={`https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(property.address)}`}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-blue-600 dark:text-blue-400 hover:underline"
                          >
                            {property.address}
                          </a>
                        ) : (
                          <div className="text-gray-900 dark:text-white">N/A</div>
                        )}
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                          ISP Company
                        </label>
                        <div className="text-gray-900 dark:text-white">{property.isp_company_name || 'N/A'}</div>
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                          ISP Account Info
                        </label>
                        <div className="text-gray-900 dark:text-white whitespace-pre-wrap">
                          {property.isp_account_info || 'N/A'}
                        </div>
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                          Notes
                        </label>
                        <div className="text-gray-900 dark:text-white whitespace-pre-wrap">
                          {property.notes || 'No notes'}
                        </div>
                      </div>
                      <div className="grid grid-cols-3 gap-4 mt-6">
                        <div className="bg-green-100 dark:bg-green-900/30 rounded-lg p-4">
                          <div className="text-2xl font-bold text-green-600 dark:text-green-400">
                            {property.online_count || 0}
                          </div>
                          <div className="text-sm text-gray-600 dark:text-gray-400">Online</div>
                        </div>
                        <div className="bg-red-100 dark:bg-red-900/30 rounded-lg p-4">
                          <div className="text-2xl font-bold text-red-600 dark:text-red-400">
                            {property.offline_count || 0}
                          </div>
                          <div className="text-sm text-gray-600 dark:text-gray-400">Offline</div>
                        </div>
                        <div className="bg-blue-100 dark:bg-blue-900/30 rounded-lg p-4">
                          <div className="text-2xl font-bold text-blue-600 dark:text-blue-400">
                            {property.total_count || 0}
                          </div>
                          <div className="text-sm text-gray-600 dark:text-gray-400">Total</div>
                        </div>
                      </div>
                    </>
                  )}
                </div>
              )}

              {activeTab === 'devices' && (
                <div>
                  <div className="mb-4 flex justify-between items-center">
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Devices</h3>
                    <button
                      onClick={syncDevices}
                      disabled={syncing}
                      className="px-4 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed"
                    >
                      {syncing ? 'Syncing...' : 'Sync from pfSense'}
                    </button>
                  </div>
                  <DevicesList
                    devices={devices}
                    propertyId={property.id}
                    onUpdate={loadDevices}
                  />
                </div>
              )}

              {activeTab === 'contacts' && (
                <ContactsList
                  contacts={contacts}
                  propertyId={property.id}
                  onUpdate={loadContacts}
                />
              )}

              {activeTab === 'attachments' && (
                <AttachmentsList
                  attachments={attachments}
                  propertyId={property.id}
                  onUpdate={loadAttachments}
                />
              )}
            </>
          )}
        </div>
      </div>
    </div>
  )
}
