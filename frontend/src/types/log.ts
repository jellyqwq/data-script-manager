export interface LogEntry {
  id: string
  script_id: string
  level: string
  message: string
  timestamp: string
}

export interface LogQuery {
  page: number
  page_size: number
  script_id?: string
  level?: string
}

export interface LogListResponse {
  data: LogEntry[]
  total: number
}
