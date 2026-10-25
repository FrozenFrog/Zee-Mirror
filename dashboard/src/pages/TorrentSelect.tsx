import { useState, useEffect, useCallback } from 'react'
import api from '../api'
import {
  FileText,
  Folder,
  RefreshCcw,
  Download,
  Check,
  X,
  AlertCircle,
  Loader2,
  ChevronDown,
  ChevronRight,
} from 'lucide-react'

interface TorrentFile {
  index: number
  path: string
  name: string
  size: number
}

interface TorrentSession {
  url?: string
  zip?: boolean
  unzip?: boolean
}

const formatBytes = (bytes: number | string): string => {
  const num = parseFloat(String(bytes))
  if (isNaN(num) || num <= 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(num) / Math.log(k))
  return parseFloat((num / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const TorrentSelect: React.FC = () => {
  const [sessionId] = useState<string | null>(() => {
    const pathParts = window.location.pathname.split('/')
    const id = pathParts[pathParts.length - 1]
    return id && id.length === 8 ? id : null
  })
  const [session, setSession] = useState<TorrentSession | null>(null)
  const [files, setFiles] = useState<TorrentFile[]>([])
  const [loading, setLoading] = useState<boolean>(!!sessionId)
  const [fileLoading, setFileLoading] = useState<boolean>(true)
  const [selectedFiles, setSelectedFiles] = useState<Set<number>>(new Set())
  const [error, setError] = useState<string | null>(!sessionId ? 'ID phiên không hợp lệ' : null)
  const [starting, setStarting] = useState<boolean>(false)
  const [success, setSuccess] = useState<boolean>(false)
  const [expandedFolders, setExpandedFolders] = useState<Set<string>>(new Set())
  const [statusLogs, setStatusLogs] = useState<string[]>([])
  const [statusMessage, setStatusMessage] = useState<string>('Đang lấy danh sách file từ Torrent...')

  const fetchSession = useCallback(async () => {
    if (!sessionId) return

    try {
      const response = await api.get(`/api/torrent/session?id=${sessionId}`)
      setSession(response.data)
      setLoading(false)
    } catch (err: any) {
      if (err.response?.status === 404) {
        setError(
          'Không tìm thấy phiên hoặc phiên đã hết hạn. Vui lòng quay lại bot và bắt đầu lại.',
        )
      } else {
        setError('Không thể tải phiên: ' + (err.response?.data || err.message))
      }
      setLoading(false)
    }
  }, [sessionId])

  const fetchFiles = useCallback(async () => {
    if (!sessionId) return

    const doFetch = async () => {
      try {
        const response = await api.get(`/api/torrent/files?id=${sessionId}`)

        if (response.data.loading) {
          setFileLoading(true)
          setStatusLogs(response.data.logs || [])
          setStatusMessage(response.data.message || 'Đang lấy metadata Torrent...')

          if (response.data.error) {
            setError(response.data.error)
            setFileLoading(false)
          } else {
            setTimeout(doFetch, 3000)
          }
        } else if (response.data.error) {
          setError(response.data.error)
          setFileLoading(false)
        } else {
          setFiles(response.data.files || [])
          const allIndices = new Set<number>((response.data.files || []).map((f: TorrentFile) => f.index))
          setSelectedFiles(allIndices)
          setFileLoading(false)
          setError(null)
        }
      } catch (err: any) {
        console.error('Failed to fetch files:', err)
        if (err.response?.status === 404) {
          setError('Không tìm thấy phiên hoặc phiên đã hết hạn.')
        } else {
          setError('Không thể lấy danh sách file: ' + (err.response?.data || err.message))
        }
        setFileLoading(false)
      }
    }

    doFetch()
  }, [sessionId])

  useEffect(() => {
    const init = async () => {
      await fetchSession()
    }
    init()
  }, [fetchSession])

  useEffect(() => {
    if (session) {
      fetchFiles()
    }
  }, [session, fetchFiles])

  const toggleFile = (index: number) => {
    const newSelected = new Set(selectedFiles)
    if (newSelected.has(index)) {
      newSelected.delete(index)
    } else {
      newSelected.add(index)
    }
    setSelectedFiles(newSelected)
  }

  const selectAll = () => {
    setSelectedFiles(new Set(files.map((f) => f.index)))
  }

  const deselectAll = () => {
    setSelectedFiles(new Set())
  }

  const handleStartDownload = async () => {
    if (selectedFiles.size === 0) {
      setError('Chọn ít nhất một file để tải xuống')
      return
    }

    setStarting(true)
    setError(null)

    try {
      await api.post(
        '/api/torrent/start',
        {
          sessionId: sessionId,
          selectedFiles: Array.from(selectedFiles),
        },
      )

      setSuccess(true)

      setTimeout(() => {
        window.close()
      }, 3000)
    } catch (err: any) {
      setError('Không thể bắt đầu tải xuống: ' + (err.response?.data || err.message))
      setStarting(false)
    }
  }

  const groupedFiles = files.reduce((acc: { [key: string]: TorrentFile[] }, file) => {
    const pathParts = file.path.split('/')
    if (pathParts.length > 1) {
      const folder = pathParts.slice(0, -1).join('/')
      if (!acc[folder]) acc[folder] = []
      acc[folder].push(file)
    } else {
      if (!acc['_root']) acc['_root'] = []
      acc['_root'].push(file)
    }
    return acc
  }, {})

  const toggleFolder = (folder: string) => {
    const newExpanded = new Set(expandedFolders)
    if (newExpanded.has(folder)) {
      newExpanded.delete(folder)
    } else {
      newExpanded.add(folder)
    }
    setExpandedFolders(newExpanded)
  }

  const selectFolder = (folder: string) => {
    const folderFiles = groupedFiles[folder] || []
    const allSelected = folderFiles.every((f) => selectedFiles.has(f.index))

    const newSelected = new Set(selectedFiles)
    folderFiles.forEach((f) => {
      if (allSelected) {
        newSelected.delete(f.index)
      } else {
        newSelected.add(f.index)
      }
    })
    setSelectedFiles(newSelected)
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-slate-50 dark:bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 flex items-center justify-center">
        <div className="text-center">
          <Loader2 size={48} className="animate-spin text-blue-500 mx-auto mb-4" />
          <p className="text-slate-600 dark:text-white/60 text-lg">Đang tải phiên Torrent...</p>
        </div>
      </div>
    )
  }

  if (error && !session) {
    return (
      <div className="min-h-screen bg-slate-50 dark:bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 flex items-center justify-center p-8">
        <div className="bg-white dark:bg-red-500/10 border border-red-200 dark:border-red-500/30 rounded-3xl p-8 max-w-md text-center shadow-2xl">
          <AlertCircle size={48} className="text-red-500 mx-auto mb-4" />
          <h2 className="text-2xl font-bold text-slate-900 dark:text-white mb-2">Lỗi</h2>
          <p className="text-slate-600 dark:text-white/60">{error}</p>
        </div>
      </div>
    )
  }

  if (success) {
    return (
      <div className="min-h-screen bg-slate-50 dark:bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 flex items-center justify-center p-8">
        <div className="bg-white dark:bg-green-500/10 border border-green-200 dark:border-green-500/30 rounded-3xl p-8 max-w-md text-center shadow-2xl">
          <Check size={48} className="text-green-500 mx-auto mb-4" />
          <h2 className="text-2xl font-bold text-slate-900 dark:text-white mb-2">
            Đã bắt đầu tải xuống!
          </h2>
          <p className="text-slate-600 dark:text-white/60">
            Quá trình tải xuống đã bắt đầu. Bạn có thể đóng trang này và quay lại bot để theo dõi
            tiến độ.
          </p>
          <p className="text-slate-400 dark:text-white/40 text-sm mt-4">
            Trang sẽ tự động đóng...
          </p>
        </div>
      </div>
    )
  }

  const totalSelected = selectedFiles.size
  const totalSize = files
    .filter((f) => selectedFiles.has(f.index))
    .reduce((sum, f) => sum + f.size, 0)

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 p-8 transition-colors duration-500">
      <div className="max-w-4xl mx-auto">
        { }
        <div className="mb-8">
          <div className="flex items-center space-x-3 mb-4">
            <div className="h-1 w-12 bg-blue-500 rounded-full" />
            <span className="text-xs font-bold tracking-widest text-blue-500 uppercase">
              Chọn file Torrent
            </span>
          </div>
          <h1 className="text-4xl font-black text-slate-900 dark:text-white mb-2">
            Chọn file để tải xuống
          </h1>
          <p className="text-slate-600 dark:text-white/60">
            Chọn các file bạn muốn tải xuống từ Torrent này
          </p>
        </div>

        { }
        {session && (
          <div className="bg-white dark:bg-white/5 backdrop-blur border border-slate-200 dark:border-white/10 rounded-2xl p-6 mb-6 shadow-sm">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs font-bold text-slate-400 dark:text-white/40 uppercase tracking-wider mb-1">
                  Liên kết Magnet
                </p>
                <p className="text-slate-700 dark:text-white/80 text-sm font-mono break-all">
                  {session.url?.slice(0, 100)}...
                </p>
              </div>
              <div className="flex items-center space-x-2">
                {session.zip && (
                  <span className="px-3 py-1 bg-purple-500/20 text-purple-600 dark:text-purple-400 rounded-full text-xs font-bold">
                    ZIP
                  </span>
                )}
                {session.unzip && (
                  <span className="px-3 py-1 bg-orange-500/20 text-orange-600 dark:text-orange-400 rounded-full text-xs font-bold">
                    UNZIP
                  </span>
                )}
              </div>
            </div>
          </div>
        )}

        { }
        {error && (
          <div className="bg-red-500/10 border border-red-500/30 rounded-xl p-4 mb-6 flex items-center space-x-3">
            <AlertCircle size={20} className="text-red-500" />
            <p className="text-red-400">{error}</p>
            <button
              onClick={() => setError(null)}
              className="ml-auto text-red-400 hover:text-red-300"
            >
              <X size={18} />
            </button>
          </div>
        )}

        { }
        <div className="bg-white dark:bg-white/5 backdrop-blur border border-slate-200 dark:border-white/10 rounded-2xl overflow-hidden mb-6 shadow-xl">
          { }
          <div className="p-4 border-b border-slate-200 dark:border-white/10 flex items-center justify-between bg-slate-50/50 dark:bg-transparent">
            <div className="flex items-center space-x-4">
              <button
                onClick={selectAll}
                className="px-4 py-2 bg-blue-500/10 dark:bg-blue-500/20 text-blue-600 dark:text-blue-400 rounded-lg text-sm font-black hover:bg-blue-500/20 dark:hover:bg-blue-500/30 transition uppercase tracking-wider"
              >
                Chọn tất cả
              </button>
              <button
                onClick={deselectAll}
                className="px-4 py-2 bg-slate-200/50 dark:bg-white/5 text-slate-600 dark:text-white/60 rounded-lg text-sm font-black hover:bg-slate-200 dark:hover:bg-white/10 transition uppercase tracking-wider"
              >
                Bỏ chọn tất cả
              </button>
            </div>
            <button
              onClick={fetchFiles}
              className="p-2 bg-slate-100 dark:bg-white/5 rounded-lg hover:bg-slate-200 dark:hover:bg-white/10 transition"
            >
              <RefreshCcw
                size={18}
                className={`text-slate-500 dark:text-white/60 ${fileLoading ? 'animate-spin' : ''}`}
              />
            </button>
          </div>

          { }
          {fileLoading ? (
            <div className="p-8">
              <div className="text-center mb-8">
                <Loader2 size={40} className="animate-spin text-blue-500 mx-auto mb-4" />
                <p className="text-slate-900 dark:text-white text-lg font-black">{statusMessage}</p>
                <p className="text-slate-500 dark:text-white/40 text-sm mt-1">
                  Trang này sẽ tự động cập nhật khi tìm thấy file.
                </p>
              </div>

              { }
              <div className="bg-slate-100 dark:bg-black/40 rounded-xl p-4 font-mono text-xs border border-slate-200 dark:border-white/5 max-h-[200px] overflow-y-auto">
                <div className="flex items-center space-x-2 mb-3 border-b border-slate-200 dark:border-white/5 pb-2">
                  <div className="w-2 h-2 rounded-full bg-blue-500 animate-pulse" />
                  <span className="text-slate-500 dark:text-white/40 uppercase tracking-widest font-bold">
                    Nhật ký tiến độ Bot
                  </span>
                </div>
                {statusLogs.length === 0 ? (
                  <p className="text-slate-400 dark:text-white/20 italic">
                    Đang chờ phản hồi từ bot...
                  </p>
                ) : (
                  statusLogs.map((log, i) => (
                    <div key={i} className="text-slate-600 dark:text-white/60 mb-1 leading-relaxed">
                      <span className="text-blue-500 dark:text-blue-400/50 mr-2">➜</span>
                      {log}
                    </div>
                  ))
                )}
              </div>

              <div className="mt-8 pt-6 border-t border-slate-200 dark:border-white/5 text-center">
                <p className="text-slate-500 dark:text-white/40 text-sm mb-4 italic">
                  &quot;Nếu Torrent có nhiều Seeders, metadata thường được lấy trong &lt; 30
                  giây. Tuy nhiên, với Torrent cũ có thể mất đến vài phút.&quot;
                </p>
                <button
                  onClick={() => handleStartDownload()}
                  className="px-6 py-2 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 text-slate-500 dark:text-white/40 rounded-lg text-xs font-black transition uppercase tracking-widest"
                >
                  Chọn tất cả & tải xuống ngay
                </button>
              </div>
            </div>
          ) : files.length === 0 ? (
            <div className="p-12 text-center">
              <AlertCircle
                size={32}
                className="text-yellow-600 dark:text-yellow-500 mx-auto mb-4"
              />
              <p className="text-slate-900 dark:text-white/60 font-black">
                Không thể lấy danh sách file
              </p>
              <p className="text-slate-500 dark:text-white/40 text-sm mt-2">
                Torrent có thể chỉ chứa một file, hoặc metadata chưa sẵn sàng
              </p>
              <button
                onClick={handleStartDownload}
                className="mt-4 px-6 py-3 bg-blue-600 text-white rounded-xl font-black uppercase tracking-widest hover:bg-blue-700 transition shadow-lg shadow-blue-500/20"
              >
                Tải xuống tất cả file
              </button>
            </div>
          ) : (
            <div className="max-h-[400px] overflow-y-auto">
              {Object.entries(groupedFiles).map(([folder, folderFiles]) => {
                const isRoot = folder === '_root'
                const isExpanded = isRoot || expandedFolders.has(folder)
                const allSelected = folderFiles.every((f) => selectedFiles.has(f.index))
                const someSelected = folderFiles.some((f) => selectedFiles.has(f.index))

                return (
                  <div key={folder}>
                    {!isRoot && (
                      <div
                        className="flex items-center px-4 py-3 bg-slate-50 dark:bg-white/[0.02] border-b border-slate-200 dark:border-white/5 cursor-pointer hover:bg-slate-100 dark:hover:bg-white/5 transition"
                        onClick={() => toggleFolder(folder)}
                      >
                        <button
                          onClick={(e) => {
                            e.stopPropagation()
                            selectFolder(folder)
                          }}
                          className={`w-5 h-5 rounded border-2 mr-3 flex items-center justify-center transition ${allSelected
                            ? 'bg-blue-500 border-blue-500'
                            : someSelected
                              ? 'bg-blue-500/50 border-blue-500'
                              : 'border-slate-300 dark:border-white/20'
                            }`}
                        >
                          {(allSelected || someSelected) && (
                            <Check size={14} className="text-white" />
                          )}
                        </button>
                        {isExpanded ? (
                          <ChevronDown
                            size={18}
                            className="text-slate-400 dark:text-white/40 mr-2"
                          />
                        ) : (
                          <ChevronRight
                            size={18}
                            className="text-slate-400 dark:text-white/40 mr-2"
                          />
                        )}
                        <Folder size={18} className="text-yellow-500 mr-3" />
                        <span className="text-slate-900 dark:text-white/80 font-black tracking-tight flex-1">
                          {folder}
                        </span>
                        <span className="text-slate-500 dark:text-white/40 text-xs font-bold uppercase">
                          {folderFiles.length} file
                        </span>
                      </div>
                    )}
                    {isExpanded &&
                      folderFiles.map((file) => (
                        <div
                          key={file.index}
                          className={`flex items-center px-4 py-3 border-b border-slate-100 dark:border-white/5 hover:bg-slate-50 dark:hover:bg-white/5 cursor-pointer transition ${!isRoot ? 'pl-12' : ''
                            }`}
                          onClick={() => toggleFile(file.index)}
                        >
                          <button
                            className={`w-5 h-5 rounded border-2 mr-3 flex items-center justify-center transition ${selectedFiles.has(file.index)
                              ? 'bg-blue-500 border-blue-500'
                              : 'border-slate-300 dark:border-white/20'
                              }`}
                          >
                            {selectedFiles.has(file.index) && (
                              <Check size={14} className="text-white" />
                            )}
                          </button>
                          <FileText size={18} className="text-blue-500 dark:text-blue-400 mr-3" />
                          <span className="text-slate-800 dark:text-white/80 flex-1 truncate font-medium">
                            {file.name}
                          </span>
                          <span className="text-slate-500 dark:text-white/40 text-xs font-black">
                            {formatBytes(file.size)}
                          </span>
                        </div>
                      ))}
                  </div>
                )
              })}
            </div>
          )}
        </div>

        { }
        {files.length > 0 && (
          <div className="bg-white dark:bg-white/5 backdrop-blur border border-slate-200 dark:border-white/10 rounded-2xl p-6 shadow-2xl">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-slate-600 dark:text-white/60 text-sm">
                  Đã chọn{' '}
                  <span className="text-slate-900 dark:text-white font-black">{totalSelected}</span>{' '}
                  /{' '}
                  <span className="text-slate-900 dark:text-white font-black">{files.length}</span>{' '}
                  file
                </p>
                <p className="text-slate-500 dark:text-white/40 text-sm mt-1">
                  Tổng dung lượng:{' '}
                  <span className="text-primary font-black">{formatBytes(totalSize)}</span>
                </p>
              </div>
              <button
                onClick={handleStartDownload}
                disabled={starting || totalSelected === 0}
                className={`flex items-center space-x-3 px-8 py-4 rounded-2xl font-black uppercase tracking-wider transition ${starting || totalSelected === 0
                  ? 'bg-slate-200 dark:bg-white/10 text-slate-400 dark:text-white/30 cursor-not-allowed'
                  : 'bg-gradient-to-r from-blue-600 to-blue-500 text-white hover:from-blue-500 hover:to-blue-400 shadow-xl shadow-blue-500/25'
                  }`}
              >
                {starting ? (
                  <>
                    <Loader2 size={20} className="animate-spin" />
                    <span>Đang bắt đầu...</span>
                  </>
                ) : (
                  <>
                    <Download size={20} />
                    <span>Tải xuống {totalSelected} file</span>
                  </>
                )}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export default TorrentSelect
