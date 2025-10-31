import { apiClient } from './api'

export interface ProcessingJob {
  id: string
  projectId: string
  type: 'georeferencing' | 'dem_generation' | '3d_model' | 'export'
  status: 'pending' | 'running' | 'completed' | 'failed'
  progress: number
  createdAt: string
  completedAt?: string
  error?: string
  result?: Record<string, any>
}

export interface StartProcessingRequest {
  projectId: string
  type: string
  parameters?: Record<string, any>
}

export const processingService = {
  startProcessing: async (request: StartProcessingRequest): Promise<ProcessingJob> => {
    const response = await apiClient.post('/api/processing/start', request)
    return response.data
  },

  getJobStatus: async (jobId: string): Promise<ProcessingJob> => {
    const response = await apiClient.get(`/api/processing/jobs/${jobId}`)
    return response.data
  },

  getProjectJobs: async (projectId: string): Promise<ProcessingJob[]> => {
    const response = await apiClient.get(`/api/processing/projects/${projectId}/jobs`)
    return response.data
  },

  cancelJob: async (jobId: string): Promise<void> => {
    await apiClient.post(`/api/processing/jobs/${jobId}/cancel`)
  },
}
