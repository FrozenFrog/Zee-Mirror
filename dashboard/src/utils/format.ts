import { isMessageKey, TFunction } from '../i18n'

export const formatBytes = (bytes: number): string => {
  const num = parseFloat(String(bytes))
  if (isNaN(num) || num <= 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(num) / Math.log(k))
  return parseFloat((num / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// Translated display label for a status value coming from the backend.
// Only used for rendering — comparisons must keep using the raw values.
export const statusLabel = (status: string, t: TFunction): string => {
  const key = `status.${status === 'canceled' ? 'cancelled' : status}`
  return isMessageKey(key) ? t(key) : status
}

export const roleLabel = (role: string, t: TFunction): string => {
  const key = `role.${role}`
  return isMessageKey(key) ? t(key) : role
}

// Dates use the active locale (useI18n().locale).
export const formatDateTime = (value: string, locale: string): string =>
  value ? new Date(value).toLocaleString(locale) : '-'

export const formatDate = (value: string, locale: string): string =>
  value ? new Date(value).toLocaleDateString(locale) : '-'

export const formatWeekday = (value: string, locale: string): string =>
  new Date(value).toLocaleDateString(locale, { weekday: 'short' })

export interface SourceInfo {
  label: string
  href?: string
}

// Short display form of a task's source link. href is only set for links a
// browser can open; torrent tasks carry a "#select=" suffix for chosen files.
export const describeSource = (raw: string, t: TFunction): SourceInfo => {
  if (!raw) return { label: '-' }
  if (raw.startsWith('tgfileid://')) return { label: t('source.telegram') }
  if (raw.startsWith('file://')) return { label: t('source.local') }
  if (raw.startsWith('magnet:')) {
    const link = raw.split('#select=')[0]
    const params = new URLSearchParams(link.slice(link.indexOf('?') + 1))
    const name = params.get('dn') || params.get('xt')?.split(':').pop() || 'magnet'
    return { label: `🧲 ${name}`, href: link }
  }
  try {
    const u = new URL(raw)
    if (u.protocol === 'http:' || u.protocol === 'https:') {
      return { label: u.host.replace(/^www\./, '') + u.pathname, href: raw }
    }
  } catch {
    // not a URL; show it as is
  }
  return { label: raw }
}
