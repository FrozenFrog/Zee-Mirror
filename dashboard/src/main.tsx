import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import './index.css'
import { PopupProvider } from './context/PopupProvider'
import { I18nProvider } from './i18n/I18nProvider'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <I18nProvider>
      <PopupProvider>
        <App />
      </PopupProvider>
    </I18nProvider>
  </React.StrictMode>,
)
