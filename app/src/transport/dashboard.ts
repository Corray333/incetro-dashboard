import axios from 'axios'
declare const Telegram: any

const api = axios.create({ baseURL: import.meta.env.VITE_API_URL, withCredentials: true })

api.interceptors.request.use(
  (config) => {
    if (Telegram && Telegram.WebApp) {
      const initData = Telegram.WebApp.initData
      if (initData) {
        config.headers.Authorization = initData
      }
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  },
)

export class DashboardTransport {
  static async getTasksOfEmployee(
    employeeUsername: string,
    period_start: number,
    period_end: number,
  ) {
    // console.log(period_start, period_end)
    const response = await api.get(`/tasks/employee/${employeeUsername}`, {
      params: { period_start, period_end },
    })
    return response.data
  }

  static async authorize() {
    await api.get(`/access`)
  }

  static async getQuarterTasks() {
    try {
      const response = await api.get(`/quarter-tasks`)
      return response.data
    } catch (error) {
      console.error(error)
    }
  }

  static async updateGoogleSheets() {
    const response = await api.post(`/update-sheets`)
    return response.data
  }

  static async updateProjectSheets(projectID: string) {
    await api.post(`/projects/${projectID}/update-sheets`)
    return
  }

  static async listProjectsWithSheets() {
    const {data} = await api.get(`/projects/with-sheets`)
    return data
  }

  static async parseMindmap(file: File) {
    const formData = new FormData()
    formData.append('file', file)

    const response = await api.post(`/mindmap`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
  }

  static async notifyAboutSalary() {
    await api.post(`/salary-notify`)
  }
}
