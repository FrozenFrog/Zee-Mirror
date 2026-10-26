import axios from 'axios'

const api = axios.create()

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('api_token')
  if (token) {
    config.headers.set('X-API-Key', token)
  }
  return config
})

// A failed OIDC sign-in: code is a stable error code (see oidcFail in
// internal/api/oidc.go) or, from older backends, a ready-made message.
export interface AuthRedirectError {
  code: string
  provider: string
}

// After an OIDC sign-in the backend redirects to /#oidc_token=… or
// /#oidc_error=<code>[&oidc_provider=<name>]. Move the token into storage,
// drop it from the URL and return the error, if any.
export const consumeAuthRedirect = (): AuthRedirectError | null => {
  const params = new URLSearchParams(window.location.hash.slice(1))
  const token = params.get('oidc_token')
  const error = params.get('oidc_error')
  if (!token && !error) return null
  window.history.replaceState(null, '', window.location.pathname + window.location.search)
  if (token) localStorage.setItem('api_token', token)
  return error ? { code: error, provider: params.get('oidc_provider') || '' } : null
}

export default api
