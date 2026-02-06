import React, { createContext, useContext, useState, useEffect } from 'react'
import { apiClient } from '../api/client'

interface User {
  id: number
  username: string
  email: string
  role: string
}

interface AuthContextType {
  user: User | null
  loading: boolean
  login: (username: string, password: string) => Promise<void>
  logout: () => void
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (token) {
      apiClient.setToken(token)
      loadUser()
    } else {
      setLoading(false)
    }
  }, [])

  const loadUser = async () => {
    try {
      const userData = await apiClient.getMe()
      setUser(userData)
    } catch (error) {
      console.error('Failed to load user:', error)
      apiClient.clearToken()
    } finally {
      setLoading(false)
    }
  }

  const login = async (username: string, password: string) => {
    const response = await apiClient.login(username, password)
    apiClient.setToken(response.token)
    setUser(response.user)
  }

  const logout = () => {
    apiClient.clearToken()
    setUser(null)
    window.location.href = '/login'
  }

  return (
    <AuthContext.Provider value={{ user, loading, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
