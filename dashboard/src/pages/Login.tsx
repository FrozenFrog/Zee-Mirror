import { useEffect, useState } from 'react'
import { ArrowRight, Bot, Activity, ShieldCheck, Lock } from 'lucide-react'
import api from '../api'

interface LoginScreenProps {
    authError: string | null
}

interface AuthConfig {
    oidc: boolean
    providerName: string
}

const Login: React.FC<LoginScreenProps> = ({ authError }) => {
    const [authConfig, setAuthConfig] = useState<AuthConfig | null>(null)

    useEffect(() => {
        api.get<AuthConfig>('/api/auth/config')
            .then((res) => setAuthConfig(res.data))
            .catch(() => setAuthConfig({ oidc: false, providerName: '' }))
    }, [])

    return (
        <div className="min-h-screen w-full bg-white dark:bg-[#050505] flex flex-col md:flex-row font-sans overflow-hidden">
            <div className="md:w-1/2 lg:w-3/5 h-[60vh] md:h-screen premium-gradient p-12 md:p-24 text-white flex flex-col relative overflow-hidden shrink-0">
                <div className="absolute inset-0 opacity-20 pointer-events-none">
                    <div className="absolute inset-0 bg-[url('https://www.transparenttextures.com/patterns/carbon-fibre.png')] opacity-10" />
                    <div className="absolute top-[-20%] right-[-20%] w-[100%] h-[100%] bg-white/10 rounded-full blur-[180px]" />
                    <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 opacity-[0.02] scale-[6] rotate-12 transition-transform duration-[10s] animate-pulse">
                        <Bot size={400} />
                    </div>
                </div>

                <div className="relative z-10 flex items-center space-x-6">
                    <div className="p-4 bg-white/10 backdrop-blur-3xl border border-white/20 rounded-3xl w-fit shadow-2xl ring-8 ring-white/5 transition-all hover:scale-105">
                        <Bot size={40} className="text-white" />
                    </div>
                    <div className="h-[1px] w-12 bg-white/20 rounded-full" />
                    <span className="text-[10px] font-black uppercase tracking-[0.8em] opacity-40">
                        Lõi hệ thống
                    </span>
                </div>

                <div className="relative z-10 my-auto py-16">
                    <div className="space-y-6">
                        <h2 className="text-[90px] lg:text-[140px] font-[1000] tracking-[-0.06em] leading-[0.75] uppercase flex flex-col">
                            <span className="animate-in fade-in slide-in-from-left-12 duration-1000">Zee</span>
                            <span className="text-white/25 italic animate-in fade-in slide-in-from-left-16 duration-1000 delay-200">
                                Mirror
                            </span>
                        </h2>
                        <div className="h-[5px] w-24 bg-primary/50 rounded-full shadow-[0_0_20px_rgba(59,130,246,0.5)]" />
                    </div>
                    <p className="text-white/80 text-xl lg:text-3xl font-bold leading-relaxed max-w-lg mt-12 opacity-80 animate-in fade-in slide-in-from-bottom-8 duration-1000 delay-500">
                        Điều phối thế hệ mới của hạ tầng đám mây bền vững và mạng lưới phân phối toàn
                        cầu.
                    </p>
                </div>

                <div className="relative z-10 grid grid-cols-1 sm:grid-cols-2 gap-10 pt-10 border-t border-white/5 mt-auto">
                    <div className="flex items-center space-x-5 group cursor-default">
                        <div className="w-16 h-16 rounded-[2rem] bg-white/5 flex items-center justify-center backdrop-blur-2xl border border-white/10 shadow-2xl shrink-0 group-hover:bg-white/10 transition-all duration-500">
                            <Activity size={28} className="text-green-400" />
                        </div>
                        <div>
                            <p className="text-[10px] font-black uppercase tracking-[0.4em] text-white/30 mb-2 leading-none">
                                Engine toàn cầu
                            </p>
                            <p className="text-sm font-black uppercase tracking-widest text-white leading-none">
                                Trạng thái đồng bộ: <span className="text-green-400">Trực tuyến</span>
                            </p>
                        </div>
                    </div>
                    <div className="flex items-center space-x-5 group cursor-default">
                        <div className="w-16 h-16 rounded-[2rem] bg-white/5 flex items-center justify-center backdrop-blur-2xl border border-white/10 shadow-2xl shrink-0 group-hover:bg-white/10 transition-all duration-500">
                            <ShieldCheck size={28} className="text-primary" />
                        </div>
                        <div>
                            <p className="text-[10px] font-black uppercase tracking-[0.4em] text-white/30 mb-2 leading-none">
                                Giao thức bảo mật
                            </p>
                            <p className="text-sm font-black uppercase tracking-widest text-white leading-none">
                                Xác thực: <span className="text-primary italic font-black">An toàn</span>
                            </p>
                        </div>
                    </div>
                </div>
            </div>

            <div className="flex-1 min-h-screen flex flex-col items-center justify-center p-12 md:p-24 relative bg-[#fcfdfe] dark:bg-zinc-950/40">
                <div className="absolute inset-0 bg-[radial-gradient(circle_at_center,_var(--tw-gradient-stops))] from-primary/10 via-transparent to-transparent opacity-40" />

                <div className="w-full max-w-sm relative z-10 -mt-12">
                    <div className="space-y-16 animate-in fade-in slide-in-from-bottom-8 duration-700">
                        <div className="space-y-8 text-center md:text-left">
                            <div className="flex items-center justify-center md:justify-start space-x-4">
                                <div className="h-[2px] w-12 bg-primary rounded-full shadow-[0_0_15px_rgba(59,130,246,0.3)]" />
                                <span className="text-[11px] font-black text-primary uppercase tracking-[0.8em]">
                                    Đăng nhập hệ thống
                                </span>
                            </div>
                            <h3 className="text-8xl font-[1000] text-slate-900 dark:text-white tracking-[-0.08em] leading-none">
                                Xin chào.
                            </h3>
                            <p className="text-slate-400 dark:text-zinc-500 font-bold text-xl leading-relaxed max-w-[320px] mx-auto md:mx-0">
                                {authConfig && !authConfig.oidc
                                    ? 'Chưa cấu hình đăng nhập. Hãy đặt OIDC_ISSUER và OIDC_CLIENT_ID trong env của bot.'
                                    : 'Hãy xác minh danh tính để kết nối.'}
                            </p>
                        </div>
                        {authError && (
                            <div
                                role="alert"
                                className="p-6 rounded-3xl bg-red-500/10 border border-red-500/20 text-red-500 font-bold text-sm leading-relaxed"
                            >
                                {authError}
                            </div>
                        )}
                        <button
                            disabled={!authConfig?.oidc}
                            onClick={() => window.location.assign('/api/auth/oidc/login')}
                            className="w-full p-10 premium-gradient rounded-[3rem] text-white font-black uppercase tracking-[0.3em] text-lg shadow-[0_32px_64px_-12px_rgba(59,130,246,0.5)] hover:shadow-primary/70 hover:scale-[1.05] active:scale-95 transition-all duration-500 flex items-center justify-center group disabled:opacity-40 disabled:pointer-events-none"
                        >
                            <span>Đăng nhập bằng {authConfig?.providerName || 'OIDC'}</span>
                            <ArrowRight
                                size={28}
                                className="ml-6 group-hover:translate-x-3 transition-transform"
                            />
                        </button>
                        <div className="flex items-center justify-center space-x-4 text-slate-300 dark:text-zinc-700">
                            <Lock size={16} />
                            <p className="text-[10px] font-black uppercase tracking-[0.6em] italic">
                                Xác thực qua OpenID Connect
                            </p>
                        </div>
                    </div>
                </div>

                <div className="absolute bottom-12 flex items-center space-x-6 opacity-30">
                    <p className="text-[10px] font-black text-slate-400 dark:text-zinc-600 uppercase tracking-[0.5em]">
                        Xây dựng cho hệ thống phân phối bền vững
                    </p>
                </div>
            </div>
        </div>
    )
}

export default Login
