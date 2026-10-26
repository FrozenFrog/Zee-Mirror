import { useState, useEffect, useCallback } from 'react'
import api from '../api'
import { User } from '../types'
import { useI18n } from './useI18n'
import type { MessageKey } from '../i18n'

export const useUsers = (apiToken: string) => {
  const { t } = useI18n()
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState<boolean>(false)
  // Kept as a key so the message follows language changes.
  const [errorKey, setErrorKey] = useState<MessageKey | null>(null)
  const error = errorKey ? t(errorKey) : null

  const fetchUsers = useCallback(async () => {
    if (!apiToken) return
    setLoading(true)
    try {
      const response = await api.get('/api/users')
      setUsers(response.data || [])
      setErrorKey(null)
    } catch {
      setErrorKey('users.loadFailed')
    } finally {
      setLoading(false)
    }
  }, [apiToken])

  const updateUser = async (userData: Record<string, unknown>): Promise<{ success: boolean; error?: string; apiKey?: string }> => {
    try {
      const response = await api.post('/api/users/update', userData)
      if (response.status >= 200 && response.status < 300) {
        await fetchUsers()
        return { success: true, apiKey: (response.data as { apiKey?: string }).apiKey }
      }
      return { success: false, error: t('common.updateFailed') }
    } catch {
      return { success: false, error: t('common.networkError') }
    }
  }

  const deleteUser = async (id: number): Promise<{ success: boolean; error?: string }> => {
    try {
      const response = await api.post('/api/users/delete', { id })
      if (response.status === 200) {
        await fetchUsers()
        return { success: true }
      }
      return { success: false, error: t('users.deleteFailed') }
    } catch {
      return { success: false, error: t('common.networkError') }
    }
  }

  const addUser = async (userData: Record<string, unknown>): Promise<{ success: boolean; error?: string; apiKey?: string }> => {
    try {
      const response = await api.post('/api/users/add', userData)
      if (response.status >= 200 && response.status < 300) {
        await fetchUsers()
        return { success: true, apiKey: (response.data as { apiKey?: string }).apiKey }
      }
      return { success: false, error: (response.data as { error?: string }).error || t('users.addFailed') }
    } catch {
      return { success: false, error: t('common.networkError') }
    }
  }

  useEffect(() => {
    fetchUsers()
  }, [fetchUsers])

  return { users, loading, error, fetchUsers, updateUser, deleteUser, addUser }
}
