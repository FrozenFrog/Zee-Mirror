import { useState, useCallback } from 'react'
import api from '../api'
import { usePopups } from './usePopups'
import { useI18n } from './useI18n'
import { FileItem } from '../types'

const useExplorer = (token: string) => {
  const [explorerPath, setExplorerPath] = useState<string>('')
  const [explorerFiles, setExplorerFiles] = useState<FileItem[]>([])
  const { showConfirm, showAlert, showToast } = usePopups()
  const { t } = useI18n()

  const fetchExplorer = useCallback(
    async (path = '') => {
      if (!token) return
      try {
        const res = await api.get(`/api/explorer/remote?path=${encodeURIComponent(path)}`)

        const normalizedData: FileItem[] = (res.data || []).map((item: Record<string, unknown>) => ({
          name: item.Name,
          displayName: item.Name,
          isDir: item.IsDir,
          size: item.Size,
          time: item.ModTime,
          status: 'cloud',
        }))

        setExplorerFiles(normalizedData)
        setExplorerPath(path)
      } catch (err) {
        console.error('Explorer error:', err)
      }
    },
    [token],
  )

  const getExternalLink = async (name: string) => {
    try {
      const path = explorerPath ? `${explorerPath}/${name}` : name
      const res = await api.get(
        `/api/explorer/remote/link?path=${encodeURIComponent(path)}`,
      )
      if (res.data.link) {
        window.open(res.data.link, '_blank')
      } else {
        showAlert(t('explorer.noLinkTitle'), t('explorer.noLinkBody'), { type: 'alert' })
      }
    } catch {
      showAlert(t('explorer.linkErrorTitle'), t('explorer.linkErrorBody'), { type: 'error' })
    }
  }

  const deleteFile = async (name: string) => {
    if (await showConfirm(t('explorer.deleteTitle'), t('explorer.deleteBody', { name }))) {
      try {
        const path = explorerPath ? `${explorerPath}/${name}` : name
        await api.delete(`/api/explorer/remote?path=${encodeURIComponent(path)}`)
        showToast(t('explorer.deleted'), 'success')
        fetchExplorer(explorerPath)
      } catch {
        showAlert(t('explorer.deleteErrorTitle'), t('explorer.deleteErrorBody'), { type: 'error' })
      }
    }
  }

  const uploadFile = async (file: File, path = '') => {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('path', path)

    try {
      await api.post('/api/explorer/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      showToast(t('explorer.uploaded', { name: file.name }), 'success')
      fetchExplorer(explorerPath)
      return { success: true }
    } catch (err: unknown) {
      showAlert(t('explorer.uploadErrorTitle'), t('explorer.uploadErrorBody', { name: file.name }), {
        type: 'error',
      })
      return { success: false, error: err }
    }
  }

  return {
    explorerPath,
    setExplorerPath,
    explorerFiles,
    fetchExplorer,
    getExternalLink,
    deleteFile,
    uploadFile,
  }
}

export default useExplorer
