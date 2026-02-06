import { useState, useEffect } from 'react'
import { apiClient } from '../api/client'
import DevicesList from './DevicesList'
import ContactsList from './ContactsList'
import AttachmentsList from './AttachmentsList'

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

  const tabs = [
    { id: 'details', label: 'Details' },
    { id: 'devices', label: 'Devices' },
    { id: 'contacts', label: 'Contacts' },
    { id: 'attachments', label: 'Attachments' },
  ]

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-6xl w-full max-h-[90vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b">
          <div>
            <h2 className="text-2xl font-bold text-gray-900">{property.name}</h2>
            {property.address && (
              <p className="text-sm text-gray-600 mt-1">{property.address}</p>
            )}
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 text-2xl font-bold"
          >
            ×
          </button>
        </div>

        {/* Tabs */}
        <div className="border-b">
          <div className="flex space-x-8 px-6">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`py-4 px-1 border-b-2 font-medium text-sm ${
                  activeTab === tab.id
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
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
              <div className="text-lg text-gray-600">Loading...</div>
            </div>
          ) : (
            <>
              {activeTab === 'details' && (
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Property Name
                    </label>
                    <div className="text-gray-900">{property.name}</div>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Address
                    </label>
                    <div className="text-gray-900">{property.address || 'N/A'}</div>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      ISP Company
                    </label>
                    <div className="text-gray-900">{property.isp_company_name || 'N/A'}</div>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      ISP Account Info
                    </label>
                    <div className="text-gray-900 whitespace-pre-wrap">
                      {property.isp_account_info || 'N/A'}
                    </div>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Notes
                    </label>
                    <div className="text-gray-900 whitespace-pre-wrap">
                      {property.notes || 'No notes'}
                    </div>
                  </div>
                  <div className="grid grid-cols-3 gap-4 mt-6">
                    <div className="bg-green-100 rounded-lg p-4">
                      <div className="text-2xl font-bold text-green-600">
                        {property.online_count || 0}
                      </div>
                      <div className="text-sm text-gray-600">Online</div>
                    </div>
                    <div className="bg-red-100 rounded-lg p-4">
                      <div className="text-2xl font-bold text-red-600">
                        {property.offline_count || 0}
                      </div>
                      <div className="text-sm text-gray-600">Offline</div>
                    </div>
                    <div className="bg-blue-100 rounded-lg p-4">
                      <div className="text-2xl font-bold text-blue-600">
                        {property.total_count || 0}
                      </div>
                      <div className="text-sm text-gray-600">Total</div>
                    </div>
                  </div>
                </div>
              )}

              {activeTab === 'devices' && (
                <DevicesList
                  devices={devices}
                  propertyId={property.id}
                  onUpdate={loadDevices}
                />
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
