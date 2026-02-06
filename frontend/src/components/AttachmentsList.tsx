import { useState } from 'react'
import { apiClient } from '../api/client'

interface AttachmentsListProps {
  attachments: any[]
  propertyId: number
  onUpdate: () => void
}

export default function AttachmentsList({ attachments, propertyId, onUpdate }: AttachmentsListProps) {
  const [showUploadModal, setShowUploadModal] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [file, setFile] = useState<File | null>(null)
  const [description, setDescription] = useState('')

  const handleUpload = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!file) return

    setUploading(true)
    try {
      await apiClient.uploadAttachment(propertyId, file, description)
      setShowUploadModal(false)
      setFile(null)
      setDescription('')
      onUpdate()
    } catch (error: any) {
      alert(error.message)
    } finally {
      setUploading(false)
    }
  }

  const handleDownload = async (attachment: any) => {
    try {
      const { url } = await apiClient.getAttachmentDownloadUrl(attachment.id)
      window.open(url, '_blank')
    } catch (error: any) {
      alert(error.message)
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this attachment?')) return
    try {
      await apiClient.deleteAttachment(id)
      onUpdate()
    } catch (error: any) {
      alert(error.message)
    }
  }

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 Bytes'
    const k = 1024
    const sizes = ['Bytes', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h3 className="text-lg font-semibold">Attachments ({attachments.length})</h3>
        <button
          onClick={() => setShowUploadModal(true)}
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
        >
          Upload File
        </button>
      </div>

      <div className="space-y-2">
        {attachments.map((attachment) => (
          <div key={attachment.id} className="bg-gray-50 rounded-lg p-4">
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <h4 className="font-medium">{attachment.filename}</h4>
                {attachment.description && (
                  <p className="text-sm text-gray-600 mt-1">{attachment.description}</p>
                )}
                <div className="flex gap-4 text-xs text-gray-500 mt-2">
                  <span>{formatFileSize(attachment.file_size)}</span>
                  <span>•</span>
                  <span>Uploaded by {attachment.uploaded_by}</span>
                  <span>•</span>
                  <span>{new Date(attachment.created_at).toLocaleDateString()}</span>
                </div>
              </div>
              <div className="flex gap-2">
                <button
                  onClick={() => handleDownload(attachment)}
                  className="px-3 py-1 text-sm bg-blue-600 text-white rounded hover:bg-blue-700"
                >
                  Download
                </button>
                <button
                  onClick={() => handleDelete(attachment.id)}
                  className="px-3 py-1 text-sm bg-red-600 text-white rounded hover:bg-red-700"
                >
                  Delete
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>

      {attachments.length === 0 && (
        <div className="text-center py-12 text-gray-500">
          No attachments uploaded for this property
        </div>
      )}

      {showUploadModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full">
            <h3 className="text-xl font-bold mb-4">Upload Attachment</h3>
            <form onSubmit={handleUpload}>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    File (Max 50MB)
                  </label>
                  <input
                    type="file"
                    onChange={(e) => setFile(e.target.files?.[0] || null)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Description (Optional)
                  </label>
                  <textarea
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                    rows={3}
                    placeholder="Add a description..."
                  />
                </div>
              </div>
              <div className="flex justify-end gap-2 mt-6">
                <button
                  type="button"
                  onClick={() => {
                    setShowUploadModal(false)
                    setFile(null)
                    setDescription('')
                  }}
                  className="px-4 py-2 bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300"
                  disabled={uploading}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:bg-gray-400"
                  disabled={uploading}
                >
                  {uploading ? 'Uploading...' : 'Upload'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
