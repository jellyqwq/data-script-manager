import request from './request'
import type { EnvGroup, EnvGroupListResponse, EnvGroupPayload } from '../types/envGroup'

export async function getEnvGroups(scriptId?: string): Promise<EnvGroupListResponse> {
  const { data } = await request.get<EnvGroupListResponse>('/auth/env-groups', {
    params: scriptId ? { script_id: scriptId } : undefined,
  })

  return {
    items: data.items ?? [],
    total: data.total ?? 0,
  }
}

export async function createEnvGroup(payload: EnvGroupPayload): Promise<EnvGroup> {
  const { data } = await request.post<EnvGroup>('/auth/env-groups', payload)
  return data
}

export async function updateEnvGroup(id: string, payload: EnvGroupPayload): Promise<void> {
  await request.put(`/auth/env-groups/${id}`, payload)
}

export async function deleteEnvGroup(id: string): Promise<void> {
  await request.delete(`/auth/env-groups/${id}`)
}
