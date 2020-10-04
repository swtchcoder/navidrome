import { fetchUtils } from 'react-admin'
import baseUrl from '../utils/baseUrl'

const url = (command, id, options) => {
  const params = new URLSearchParams()
  params.append('u', localStorage.getItem('username'))
  params.append('t', localStorage.getItem('subsonic-token'))
  params.append('s', localStorage.getItem('subsonic-salt'))
  params.append('f', 'json')
  params.append('v', '1.8.0')
  params.append('c', 'NavidromeUI')
  params.append('id', id)
  if (options) {
    if (options.ts) {
      options['_'] = new Date().getTime()
      delete options.ts
    }
    Object.keys(options).forEach((k) => {
      params.append(k, options[k])
    })
  }
  const url = `/rest/${command}?${params.toString()}`
  return baseUrl(url)
}

const scrobble = (id, submit) =>
  fetchUtils.fetchJson(url('scrobble', id, { submission: submit }))

const star = (id) => fetchUtils.fetchJson(url('star', id))

const unstar = (id) => fetchUtils.fetchJson(url('unstar', id))

const download = (id) => (window.location.href = url('download', id))

export default { url, scrobble, download, star, unstar }
