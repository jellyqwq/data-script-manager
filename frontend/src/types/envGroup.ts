export interface EnvPair {
  key: string
  value: string
}

export interface EnvGroup {
  id: string
  script_id: string
  name: string
  enabled: boolean
  vars: EnvPair[]
  created_at?: string
  updated_at?: string
}

export interface EnvGroupListResponse {
  items: EnvGroup[]
  total: number
}

export interface EnvGroupPayload {
  script_id: string
  name: string
  enabled: boolean
  vars: EnvPair[]
}
