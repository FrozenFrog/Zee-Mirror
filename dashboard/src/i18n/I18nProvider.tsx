import { useCallback, useEffect, useMemo, useState, ReactNode } from 'react'
import { I18nContext } from './I18nContext'
import { detectLang, localeOf, saveLang, translate, Lang, MessageKey, TranslateVars } from './index'

export const I18nProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [lang, setLangState] = useState<Lang>(detectLang)

  useEffect(() => {
    document.documentElement.lang = lang
  }, [lang])

  const setLang = useCallback((next: Lang) => {
    saveLang(next)
    setLangState(next)
  }, [])

  const value = useMemo(
    () => ({
      lang,
      locale: localeOf(lang),
      setLang,
      t: (key: MessageKey, vars?: TranslateVars) => translate(lang, key, vars),
    }),
    [lang, setLang],
  )

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>
}
