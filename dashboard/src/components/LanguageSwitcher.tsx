import { Languages } from 'lucide-react'
import { useI18n } from '../hooks/useI18n'
import { LANGUAGES, isLang } from '../i18n'
import { cn } from '../utils/cn'

const LanguageSwitcher: React.FC<{ className?: string }> = ({ className }) => {
  const { lang, setLang, t } = useI18n()

  return (
    <label
      className={cn(
        'flex items-center space-x-3 px-4 py-3 rounded-2xl bg-slate-50/50 dark:bg-zinc-900/40 border border-slate-100 dark:border-white/5 text-slate-500 dark:text-zinc-400 focus-within:ring-2 ring-primary/30 transition-all',
        className,
      )}
    >
      <Languages size={16} className="shrink-0 text-primary" aria-hidden />
      <select
        value={lang}
        onChange={(e) => isLang(e.target.value) && setLang(e.target.value)}
        aria-label={t('common.language')}
        title={t('common.language')}
        className="flex-1 min-w-0 bg-transparent outline-none text-xs font-bold cursor-pointer text-slate-700 dark:text-zinc-200"
      >
        {LANGUAGES.map((l) => (
          <option key={l.code} value={l.code}>
            {l.flag} {l.name}
          </option>
        ))}
      </select>
    </label>
  )
}

export default LanguageSwitcher
