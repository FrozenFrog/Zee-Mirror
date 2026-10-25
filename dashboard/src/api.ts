import axios from 'axios'

const api = axios.create()

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('api_token')
  if (token) {
    config.headers.set('X-API-Key', token)
  }
  return config
})

// After an OIDC sign-in the backend redirects to /#oidc_token=… or
// /#oidc_error=…. Move the token into storage, drop it from the URL and
// return the error, if any.
export const consumeAuthRedirect = (): string | null => {
  const params = new URLSearchParams(window.location.hash.slice(1))
  const token = params.get('oidc_token')
  const error = params.get('oidc_error')
  if (!token && !error) return null
  window.history.replaceState(null, '', window.location.pathname + window.location.search)
  if (token) localStorage.setItem('api_token', token)
  return error
}

export default api
