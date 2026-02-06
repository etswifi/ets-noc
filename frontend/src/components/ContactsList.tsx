import { useState } from 'react'
import { apiClient } from '../api/client'

interface ContactsListProps {
  contacts: any[]
  propertyId: number
  onUpdate: () => void
}

export default function ContactsList({ contacts, propertyId, onUpdate }: ContactsListProps) {
  const [showAddModal, setShowAddModal] = useState(false)
  const [editingContact, setEditingContact] = useState<any>(null)
  const [formData, setFormData] = useState({
    name: '',
    phone: '',
    email: '',
    role: '',
    notes: '',
  })

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      if (editingContact) {
        await apiClient.updateContact(editingContact.id, formData)
      } else {
        await apiClient.createContact(propertyId, formData)
      }
      setShowAddModal(false)
      setEditingContact(null)
      setFormData({ name: '', phone: '', email: '', role: '', notes: '' })
      onUpdate()
    } catch (error: any) {
      alert(error.message)
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this contact?')) return
    try {
      await apiClient.deleteContact(id)
      onUpdate()
    } catch (error: any) {
      alert(error.message)
    }
  }

  const openEditModal = (contact: any) => {
    setEditingContact(contact)
    setFormData({
      name: contact.name,
      phone: contact.phone || '',
      email: contact.email || '',
      role: contact.role || '',
      notes: contact.notes || '',
    })
    setShowAddModal(true)
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h3 className="text-lg font-semibold">Contacts ({contacts.length})</h3>
        <button
          onClick={() => setShowAddModal(true)}
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
        >
          Add Contact
        </button>
      </div>

      <div className="space-y-2">
        {contacts.map((contact) => (
          <div key={contact.id} className="bg-gray-50 rounded-lg p-4">
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <div className="flex items-center gap-2">
                  <h4 className="font-medium">{contact.name}</h4>
                  {contact.role && (
                    <span className="text-xs bg-blue-100 text-blue-700 px-2 py-1 rounded-full">
                      {contact.role}
                    </span>
                  )}
                </div>
                {contact.phone && (
                  <div className="text-sm text-gray-600 mt-1">📞 {contact.phone}</div>
                )}
                {contact.email && (
                  <div className="text-sm text-gray-600">✉️ {contact.email}</div>
                )}
                {contact.notes && (
                  <div className="text-sm text-gray-500 mt-2">{contact.notes}</div>
                )}
              </div>
              <div className="flex gap-2">
                <button
                  onClick={() => openEditModal(contact)}
                  className="px-3 py-1 text-sm bg-gray-200 text-gray-700 rounded hover:bg-gray-300"
                >
                  Edit
                </button>
                <button
                  onClick={() => handleDelete(contact.id)}
                  className="px-3 py-1 text-sm bg-red-600 text-white rounded hover:bg-red-700"
                >
                  Delete
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>

      {contacts.length === 0 && (
        <div className="text-center py-12 text-gray-500">
          No contacts added for this property
        </div>
      )}

      {showAddModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full">
            <h3 className="text-xl font-bold mb-4">
              {editingContact ? 'Edit Contact' : 'Add Contact'}
            </h3>
            <form onSubmit={handleSubmit}>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Name *</label>
                  <input
                    type="text"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Phone</label>
                  <input
                    type="tel"
                    value={formData.phone}
                    onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Email</label>
                  <input
                    type="email"
                    value={formData.email}
                    onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Role</label>
                  <input
                    type="text"
                    value={formData.role}
                    onChange={(e) => setFormData({ ...formData, role: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                    placeholder="e.g., Manager, IT Contact"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Notes</label>
                  <textarea
                    value={formData.notes}
                    onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                    rows={3}
                  />
                </div>
              </div>
              <div className="flex justify-end gap-2 mt-6">
                <button
                  type="button"
                  onClick={() => {
                    setShowAddModal(false)
                    setEditingContact(null)
                    setFormData({ name: '', phone: '', email: '', role: '', notes: '' })
                  }}
                  className="px-4 py-2 bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
                >
                  {editingContact ? 'Update' : 'Add'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
