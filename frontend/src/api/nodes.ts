import request from './request'

export function fetchNodes() {
  return request.get('/auth/nodes')
}
