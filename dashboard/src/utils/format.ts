export const formatBytes = (bytes: number): string => {
  const num = parseFloat(String(bytes))
  if (isNaN(num) || num <= 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(num) / Math.log(k))
  return parseFloat((num / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// Vietnamese display labels for status values coming from the backend.
// Only used for rendering — comparisons must keep using the raw values.
const STATUS_LABELS: Record<string, string> = {
  all: 'Tất cả',
  downloading: 'Đang tải xuống',
  uploading: 'Đang tải lên',
  queued: 'Đang chờ',
  pending: 'Chờ xử lý',
  completed: 'Hoàn tất',
  failed: 'Thất bại',
  cancelled: 'Đã hủy',
  canceled: 'Đã hủy',
  active: 'Đang chạy',
  banned: 'Bị cấm',
}

export const statusLabel = (status: string): string => STATUS_LABELS[status] ?? status

const ROLE_LABELS: Record<string, string> = {
  owner: 'Chủ sở hữu',
  admin: 'Quản trị viên',
  authorized: 'Được cấp quyền',
  user: 'Người dùng',
}

export const roleLabel = (role: string): string => ROLE_LABELS[role] ?? role

export interface SourceInfo {
  label: string
  href?: string
}

// Short display form of a task's source link. href is only set for links a
// browser can open; torrent tasks carry a "#select=" suffix for chosen files.
export const describeSource = (raw: string): SourceInfo => {
  if (!raw) return { label: '-' }
  if (raw.startsWith('tgfileid://')) return { label: 'File Telegram' }
  if (raw.startsWith('file://')) return { label: 'File cục bộ' }
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
