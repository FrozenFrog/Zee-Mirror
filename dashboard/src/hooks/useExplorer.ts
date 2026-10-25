import { useState, useCallback } from 'react'
import api from '../api'
import { usePopups } from './usePopups'
import { FileItem } from '../types'

const useExplorer = (token: string) => {
  const [explorerPath, setExplorerPath] = useState<string>('')
  const [explorerFiles, setExplorerFiles] = useState<FileItem[]>([])
  const { showConfirm, showAlert, showToast } = usePopups()

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
        showAlert(
          'Không có liên kết',
          'Hiện chưa có liên kết công khai cho tài nguyên này.',
          { type: 'alert' },
        )
      }
    } catch {
      showAlert('Lỗi tạo liên kết', 'Không thể tạo liên kết đám mây bảo mật cho tài nguyên này.', {
        type: 'error',
      })
    }
  }

  const deleteFile = async (name: string) => {
    if (
      await showConfirm(
        'Xóa tài nguyên',
        `Xóa vĩnh viễn ${name}? Dữ liệu sẽ bị xóa khỏi kho lưu trữ đám mây.`,
      )
    ) {
      try {
        const path = explorerPath ? `${explorerPath}/${name}` : name
        await api.delete(`/api/explorer/remote?path=${encodeURIComponent(path)}`)
        showToast('Đã xóa tài nguyên thành công', 'success')
        fetchExplorer(explorerPath)
      } catch {
        showAlert(
          'Lỗi khi xóa',
          'Không thể xóa mục đã chọn khỏi kho lưu trữ đám mây.',
          { type: 'error' },
        )
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
      showToast(`Đã tải lên ${file.name} thành công`, 'success')
      fetchExplorer(explorerPath)
      return { success: true }
    } catch (err: unknown) {
      showAlert('Lỗi tải lên', `Không thể tải ${file.name} lên kho lưu trữ từ xa.`, {
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
