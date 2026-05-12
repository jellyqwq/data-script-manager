import request from './request'
import type { LogListResponse, LogQuery } from '../types/log'

export async function getLogs(params: LogQuery): Promise<LogListResponse> {
  const { data } = await request.get<LogListResponse>('/auth/logs', { params })
  return {
    data: data.data ?? [],
    total: data.total ?? 0,
  }
}

export async function deleteLog(id: string): Promise<void> {
  await request.delete(`/auth/logs/${id}`)
}

export async function clearLogs(): Promise<void> {
  await request.delete('/auth/logs')
}
