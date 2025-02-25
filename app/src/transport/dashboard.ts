import axios from "axios"

const api = axios.create({ baseURL: import.meta.env.VITE_API_URL,   withCredentials: true})

export class DashboardTransport {
    static async getTasksOfEmployee(employeeUsername: string, period_start: number, period_end: number) {
        console.log(period_start, period_end)
        const response = await api.get(`/tasks/employee/${employeeUsername}`, { params: { period_start, period_end } })
        return response.data
    }

    static async getQuarterTasks() {
        try {
            const response = await api.get(`/quarter-tasks`)
            return response.data
        
        } catch (error) {
            console.log(error)
        }
    }

    static async updateGoogleSheets(){
        const response = await api.post(`/update-sheets`)
        return response.data
    }

    static async parseMindmap(file: File) {
        const formData = new FormData();
        formData.append('file', file);

        const response = await api.post(`/mindmap`, formData, {
            headers: {
                'Content-Type': 'multipart/form-data'
            }
        });
    }

    static async notifyAboutSalary () {
        await api.post(`/salary-notify`)
    }
}