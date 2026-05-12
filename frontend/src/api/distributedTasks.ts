import request from './request'

export function submitDistributedTask(data: unknown) {
  return request.post('/distributed-tasks', data)
}

export function fetchDistributedTasks() {
  return request.get('/distributed-tasks')
}
