import vi, { type MessageKey, type Messages } from './vi'
import en from './en'
import id from './id'
import ja from './ja'
import zh from './zh'

export type { MessageKey, Messages }

// Same set and order as the bot (pkg/i18n.Languages).
export const LANGUAGES = [
  { code: 'vi', flag: '🇻🇳', name: 'Tiếng Việt', locale: 'vi-VN' },
  { code: 'en', flag: '🇺🇸', name: 'English', locale: 'en-US' },
  { code: 'id', flag: '🇮🇩', name: 'Bahasa Indonesia', locale: 'id-ID' },
  { code: 'ja', flag: '🇯🇵', name: '日本語', locale: 'ja-JP' },
  { code: 'zh', flag: '🇨🇳', name: '简体中文', locale: 'zh-CN' },
] as const

export type Lang = (typeof LANGUAGES)[number]['code']
export type TranslateVars = Record<string, string | number>
export type TFunction = (key: MessageKey, vars?: TranslateVars) => string

export const DEFAULT_LANG: Lang = 'vi'
const STORAGE_KEY = 'language'

const dictionaries: Record<Lang, Messages> = { vi, en, id, ja, zh }

export const isLang = (code: string): code is Lang => code in dictionaries

export const isMessageKey = (key: string): key is MessageKey =>
  Object.prototype.hasOwnProperty.call(vi, key)

// "en-US" -> "en", "zh-Hant-TW" -> "zh"; null when unsupported.
export const normalizeLang = (code: string | null | undefined): Lang | null => {
  if (!code) return null
  const base = code.toLowerCase().split(/[-_]/)[0]
  return isLang(base) ? base : null
}

export const localeOf = (lang: Lang): string =>
  LANGUAGES.find((l) => l.code === lang)?.locale ?? 'vi-VN'

// Saved choice first, then the browser's preferred languages, else 'vi'.
export const detectLang = (): Lang => {
  try {
    const saved = normalizeLang(localStorage.getItem(STORAGE_KEY))
    if (saved) return saved
  } catch {
    // storage unavailable (private mode, blocked cookies)
  }
  const preferred =
    typeof navigator === 'undefined'
      ? []
      : navigator.languages?.length
        ? navigator.languages
        : [navigator.language]
  for (const code of preferred) {
    const lang = normalizeLang(code)
    if (lang) return lang
  }
  return DEFAULT_LANG
}

export const saveLang = (lang: Lang) => {
  try {
    localStorage.setItem(STORAGE_KEY, lang)
  } catch {
    // ignore: the choice just won't persist
  }
}

// Replaces {name} placeholders; unknown placeholders are left as is.
export const interpolate = (template: string, vars?: TranslateVars): string =>
  vars
    ? template.replace(/\{(\w+)\}/g, (match, name: string) =>
        name in vars ? String(vars[name]) : match,
      )
    : template

export const translate = (lang: Lang, key: MessageKey, vars?: TranslateVars): string =>
  interpolate(dictionaries[lang][key] ?? vi[key] ?? key, vars)
