export interface Script {
  id: string
  script_name: string
  description?: string
  content?: string
  original_filename?: string
  file_path?: string
  sha1?: string
  size?: number
  language?: string
  created_at?: string
  last_modified?: string
}

export interface ScriptListResponse {
  items: Script[]
  total: number
}
