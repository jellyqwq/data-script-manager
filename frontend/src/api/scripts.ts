import request from './request'
import type { Script, ScriptListResponse } from '../types/script'

export interface ScriptListParams {
  page?: number
  pageSize?: number
}

export interface ScriptPayload {
  scriptName: string
  description: string
  file?: File | null
}

export async function getScripts(params: ScriptListParams = {}): Promise<ScriptListResponse> {
  const { data } = await request.get<ScriptListResponse | Script[]>('/auth/scripts', { params })

  if (Array.isArray(data)) {
    return {
      items: data,
      total: data.length,
    }
  }

  return {
    items: data.items ?? [],
    total: data.total ?? 0,
  }
}

export async function createScript(payload: ScriptPayload): Promise<Script> {
  const formData = buildScriptFormData(payload)
  const { data } = await request.post<Script>('/auth/scripts', formData)
  return data
}

export async function updateScript(id: string, payload: ScriptPayload): Promise<void> {
  const formData = buildScriptFormData(payload)
  await request.put(`/auth/scripts/${id}`, formData)
}

export async function deleteScript(id: string): Promise<void> {
  await request.delete(`/auth/scripts/${id}`)
}

function buildScriptFormData(payload: ScriptPayload) {
  const formData = new FormData()
  formData.append('script_name', payload.scriptName)
  formData.append('description', payload.description)
  if (payload.file) {
    formData.append('file', payload.file)
  }
  return formData
}
