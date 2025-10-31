import { apiClient } from './api'

export interface Project {
  id: string
  name: string
  description?: string
  status: 'draft' | 'processing' | 'completed' | 'failed'
  createdAt: string
  updatedAt: string
  metadata?: Record<string, any>
}

export interface CreateProjectRequest {
  name: string
  description?: string
}

export const projectService = {
  getProjects: async (): Promise<Project[]> => {
    const response = await apiClient.get('/api/projects')
    return response.data
  },

  getProject: async (id: string): Promise<Project> => {
    const response = await apiClient.get(`/api/projects/${id}`)
    return response.data
  },

  createProject: async (project: CreateProjectRequest): Promise<Project> => {
    const response = await apiClient.post('/api/projects', project)
    return response.data
  },

  updateProject: async (id: string, updates: Partial<Project>): Promise<Project> => {
    const response = await apiClient.put(`/api/projects/${id}`, updates)
    return response.data
  },

  deleteProject: async (id: string): Promise<void> => {
    await apiClient.delete(`/api/projects/${id}`)
  },
}
