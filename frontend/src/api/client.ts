const API_BASE_URL = import.meta.env.VITE_API_URL || ''

class ApiClient {
  private baseUrl: string
  private token: string | null = null

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl
    this.token = localStorage.getItem('token')
  }

  setToken(token: string) {
    this.token = token
    localStorage.setItem('token', token)
  }

  clearToken() {
    this.token = null
    localStorage.removeItem('token')
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string>),
    }

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`
    }

    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      ...options,
      headers,
    })

    if (response.status === 401) {
      this.clearToken()
      window.location.href = '/login'
      throw new Error('Unauthorized')
    }

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Request failed' }))
      throw new Error(error.error || 'Request failed')
    }

    return response.json()
  }

  // Auth
  async login(username: string, password: string) {
    return this.request<{ token: string; user: any }>('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    })
  }

  async getMe() {
    return this.request<any>('/api/v1/auth/me')
  }

  // Dashboard
  async getDashboard() {
    return this.request<any>('/api/v1/dashboard')
  }

  // Properties
  async getProperties() {
    return this.request<any[]>('/api/v1/properties')
  }

  async getProperty(id: number) {
    return this.request<any>(`/api/v1/properties/${id}`)
  }

  async createProperty(data: any) {
    return this.request<any>('/api/v1/properties', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async updateProperty(id: number, data: any) {
    return this.request<any>(`/api/v1/properties/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    })
  }

  async deleteProperty(id: number) {
    return this.request<any>(`/api/v1/properties/${id}`, {
      method: 'DELETE',
    })
  }

  async getPropertyStatus(id: number) {
    return this.request<any>(`/api/v1/properties/${id}/status`)
  }

  async getPropertyDevices(id: number) {
    return this.request<any[]>(`/api/v1/properties/${id}/devices`)
  }

  async syncDevicesFromPfSense(id: number) {
    return this.request<any>(`/api/v1/properties/${id}/sync-devices`, {
      method: 'POST',
    })
  }

  // Contacts
  async getContacts(propertyId: number) {
    return this.request<any[]>(`/api/v1/properties/${propertyId}/contacts`)
  }

  async createContact(propertyId: number, data: any) {
    return this.request<any>(`/api/v1/properties/${propertyId}/contacts`, {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async updateContact(id: number, data: any) {
    return this.request<any>(`/api/v1/contacts/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    })
  }

  async deleteContact(id: number) {
    return this.request<any>(`/api/v1/contacts/${id}`, {
      method: 'DELETE',
    })
  }

  // Attachments
  async getAttachments(propertyId: number) {
    return this.request<any[]>(`/api/v1/properties/${propertyId}/attachments`)
  }

  async uploadAttachment(propertyId: number, file: File, description: string) {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('description', description)

    const headers: HeadersInit = {}
    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`
    }

    const response = await fetch(
      `${this.baseUrl}/api/v1/properties/${propertyId}/attachments`,
      {
        method: 'POST',
        headers,
        body: formData,
      }
    )

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Upload failed' }))
      throw new Error(error.error || 'Upload failed')
    }

    return response.json()
  }

  async getAttachmentDownloadUrl(id: number) {
    return this.request<{ url: string }>(`/api/v1/attachments/${id}/download`)
  }

  async deleteAttachment(id: number) {
    return this.request<any>(`/api/v1/attachments/${id}`, {
      method: 'DELETE',
    })
  }

  // Devices
  async getDevices() {
    return this.request<any[]>('/api/v1/devices')
  }

  async getDevice(id: number) {
    return this.request<any>(`/api/v1/devices/${id}`)
  }

  async createDevice(data: any) {
    return this.request<any>('/api/v1/devices', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async updateDevice(id: number, data: any) {
    return this.request<any>(`/api/v1/devices/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    })
  }

  async deleteDevice(id: number) {
    return this.request<any>(`/api/v1/devices/${id}`, {
      method: 'DELETE',
    })
  }

  async getDeviceStatus(id: number) {
    return this.request<any>(`/api/v1/devices/${id}/status`)
  }

  async getDeviceHistory(id: number, start?: string, end?: string) {
    let url = `/api/v1/devices/${id}/history`
    const params = new URLSearchParams()
    if (start) params.append('start', start)
    if (end) params.append('end', end)
    if (params.toString()) url += `?${params.toString()}`
    return this.request<any[]>(url)
  }

  async getDeviceErrors(id: number, limit: number = 10) {
    return this.request<any[]>(`/api/v1/devices/${id}/errors?limit=${limit}`)
  }

  // Users
  async getUsers() {
    return this.request<any[]>('/api/v1/users')
  }

  async createUser(data: any) {
    return this.request<any>('/api/v1/users', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async updateUser(id: number, data: any) {
    return this.request<any>(`/api/v1/users/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    })
  }

  async deleteUser(id: number) {
    return this.request<any>(`/api/v1/users/${id}`, {
      method: 'DELETE',
    })
  }

  // Settings
  async getSettings() {
    return this.request<any>('/api/v1/settings')
  }

  async updateSettings(data: any) {
    return this.request<any>('/api/v1/settings', {
      method: 'PUT',
      body: JSON.stringify(data),
    })
  }
}

export const apiClient = new ApiClient(API_BASE_URL)
