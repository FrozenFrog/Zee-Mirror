import { createContext } from 'react'
import type { Lang, TFunction } from './index'

export interface I18nContextType {
  lang: Lang
  locale: string
  setLang: (lang: Lang) => void
  t: TFunction
}

export const I18nContext = createContext<I18nContextType | null>(null)
