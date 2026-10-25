import { useState, useEffect } from 'react'
import api from '../api'
import {
  Activity,
  Download,
  HardDrive,
  ShieldCheck,
  TrendingUp,
  Clock,
  Layers,
  Zap,
} from 'lucide-react'
import StatsCard from '../components/Stats/StatsCard'
import TaskRow from '../components/Task/TaskRow'
import HealthBar from '../components/Stats/HealthBar'
import StatusItem from '../components/Stats/StatusItem'
import { formatBytes } from '../utils/format'
import { Task, Stats, SystemMetrics, User } from '../types'

interface OverviewProps {
  tasks: Task[]
  stats: Stats
  system: SystemMetrics
  onCancelTask: (id: string) => void
  setActiveTab: (tab: string) => void
  apiToken: string
}

const Overview: React.FC<OverviewProps> = ({ tasks, stats, system, onCancelTask, setActiveTab, apiToken }) => {
  const [users, setUsers] = useState<User[]>([])
  useEffect(() => {
    if (!apiToken) return
    api.get('/api/users')
      .then(res => setUsers(res.data || []))
      .catch(() => {})
  }, [apiToken])
  return (
    <div className="space-y-16 relative z-10 animate-in fade-in zoom-in-95 duration-1000">
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8">
        <StatsCard
          icon={TrendingUp}
          label="Đang xử lý"
          value={tasks.length}
          subLabel="Task đang hoạt động"
          color="primary"
        />
        <StatsCard
          icon={Download}
          label="Đã hoàn tất"
          value={stats.total_tasks || 0}
          subLabel="Tổng Task thành công"
          color="blue"
        />
        <StatsCard
          icon={HardDrive}
          label="Lưu lượng"
          value={formatBytes(stats.total_bandwidth || 0)}
          subLabel="Tổng dữ liệu truyền tải"
          color="indigo"
        />
        <StatsCard
          icon={ShieldCheck}
          label="Được cấp quyền"
          value={stats.users_count || 1}
          subLabel="Người dùng được phép"
          color="green"
        />
      </div>

      {/* Quota summary card */}
      <div className="bg-white/80 dark:bg-zinc-900/60 backdrop-blur-xl rounded-3xl border border-slate-200 dark:border-white/5 shadow-2xl shadow-black/5 p-8">
        <h3 className="text-lg font-black text-slate-900 dark:text-white tracking-tight mb-6">
          Hạn mức sử dụng hôm nay
        </h3>
        <div className="space-y-4">
          {users.filter(u => u.maxDailyTasks !== -1 || u.maxDailyBandwidth !== -1).length > 0 ? (
            users.filter(u => u.maxDailyTasks !== -1 || u.maxDailyBandwidth !== -1).map((user) => (
              <div key={user.id} className="space-y-2">
                <div className="flex justify-between text-xs font-bold">
                  <span className="text-slate-600 dark:text-slate-300">{user.username || `Người dùng ${user.id}`}</span>
                  <span className="text-slate-400">
                    Task: {user.usedTasks ?? 0}/{user.maxDailyTasks === -1 ? '∞' : user.maxDailyTasks}
                    {user.maxDailyBandwidth !== -1 && ` • Băng thông: ${formatBytes(user.usedBandwidth ?? 0)}/${formatBytes(user.maxDailyBandwidth)}`}
                  </span>
                </div>
                {user.maxDailyTasks !== -1 && (
                  <div className="w-full h-2 bg-slate-200 dark:bg-white/10 rounded-full overflow-hidden">
                    <div className="h-full bg-primary rounded-full transition-all" style={{ width: `${Math.min(100, ((user.usedTasks ?? 0) / user.maxDailyTasks) * 100)}%` }} />
                  </div>
                )}
              </div>
            ))
          ) : (
            <p className="text-sm text-slate-400">Tất cả người dùng đều có hạn mức không giới hạn.</p>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 xl:grid-cols-3 gap-12">
        <div className="xl:col-span-2 space-y-10">
          <div className="flex justify-between items-end px-2">
            <div className="flex items-center space-x-4">
              <div className="p-3 bg-primary/10 rounded-[1.25rem] text-primary">
                <Clock size={24} />
              </div>
              <h3 className="text-3xl font-black text-slate-900 dark:text-white tracking-tight">
                Tiến trình đang chạy
              </h3>
            </div>
            <button
              onClick={() => setActiveTab('tasks')}
              className="px-6 py-2.5 bg-slate-100 dark:bg-zinc-900/60 rounded-xl text-[9px] font-black uppercase tracking-widest text-primary hover:bg-primary hover:text-white transition-all shadow-sm border border-slate-200 dark:border-white/5"
            >
              Xem tất cả
            </button>
          </div>
          <div className="space-y-6">
            {tasks.length > 0 ? (
              <div className="grid grid-cols-1 gap-6">
                {tasks.slice(0, 3).map((task) => (
                  <TaskRow
                    key={task.id}
                    task={task}
                    formatBytes={formatBytes}
                    onCancel={onCancelTask}
                  />
                ))}
              </div>
            ) : (
              <div className="glass-card py-32 rounded-[3.5rem] flex flex-col items-center justify-center text-center group border-dashed border-2 border-slate-200 dark:border-white/10">
                <div className="p-12 bg-slate-100 dark:bg-white/5 rounded-full mb-8 group-hover:scale-110 transition-all duration-700 shadow-inner border border-slate-200 dark:border-transparent">
                  <Activity size={72} className="text-slate-400 dark:text-zinc-800 opacity-60" />
                </div>
                <h4 className="text-2xl font-black text-slate-600 dark:text-zinc-600 uppercase tracking-[0.2em] leading-relaxed">
                  Cụm đang rảnh
                  <br />
                  <span className="text-[10px] text-primary/60 tracking-[0.5em]">
                    Đang chờ Task mới
                  </span>
                </h4>
              </div>
            )}
          </div>
        </div>

        <div className="space-y-12">
          <div className="flex items-center space-x-4 px-2">
            <div className="p-3 bg-indigo-500/10 rounded-[1.25rem] text-indigo-400">
              <Layers size={24} />
            </div>
            <h3 className="text-3xl font-black text-slate-900 dark:text-white tracking-tight">
              Chỉ số Node
            </h3>
          </div>
          <div className="glass-card p-12 rounded-[4rem] space-y-12 relative overflow-hidden group">
            <div className="absolute top-0 right-0 p-8 opacity-[0.03] group-hover:opacity-[0.07] transition-all duration-1000 rotate-12">
              <Zap size={140} />
            </div>
            <div className="space-y-12 relative z-10">
              <HealthBar
                label="Tải CPU"
                value={Math.round(system.cpu || 0)}
                color="from-primary to-blue-400"
              />
              <HealthBar
                label="Sử dụng RAM"
                value={Math.round(system.ram || 0)}
                color="from-indigo-500 to-purple-400"
              />
              <HealthBar
                label="Dung lượng lưu trữ"
                value={Math.round(system.disk || 0)}
                color="from-blue-600 to-cyan-400"
              />
            </div>

            <div className="pt-12 border-t border-slate-100 dark:border-white/5 space-y-7 relative z-10">
              <StatusItem label="Dịch vụ" value="Đang hoạt động" dot="bg-green-500" />
              <StatusItem
                label="Kiến trúc"
                value={`${system.os || 'Linux'} ${system.arch || 'x64'}`}
                color="text-slate-600"
              />
              <StatusItem
                label="Uptime"
                value={
                  system.uptime
                    ? `${Math.floor(system.uptime / 3600)} giờ ${Math.floor((system.uptime % 3600) / 60)} phút`
                    : 'Ổn định'
                }
                color="text-primary"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default Overview
