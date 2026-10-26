import { useState, useCallback } from 'react'
import api from '../api'
import { usePopups } from './usePopups'
import { useI18n } from './useI18n'
import { Task } from '../types'

const useTasks = (token: string) => {
  const [tasks, setTasks] = useState<Task[]>([])
  const { showConfirm, showToast } = usePopups()
  const { t } = useI18n()

  const fetchTasks = useCallback(async () => {
    if (!token) return
    try {
      const res = await api.get('/api/tasks')
      setTasks(res.data || [])
    } catch (err) {
      console.error(err)
    }
  }, [token])

  const cancelTask = async (id: string) => {
    if (await showConfirm(t('tasks.cancelTitle'), t('tasks.cancelBody'))) {
      try {
        await api.delete(`/api/tasks?id=${id}`)
        showToast(t('tasks.cancelled'), 'success')
        fetchTasks()
      } catch (err) {
        console.error(err)
        showToast(t('tasks.cancelFailed'), 'error')
      }
    }
  }

  return { tasks, fetchTasks, cancelTask, setTasks }
}

export default useTasks
