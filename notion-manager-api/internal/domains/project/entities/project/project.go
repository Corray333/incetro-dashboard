package project

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID         uuid.UUID `json:"id"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	Name       string    `json:"name"`
	Icon       string    `json:"icon"`
	IconType   string    `json:"iconType"`
	Status     string    `json:"status"`
	Type       string    `json:"type"`
	SheetsLink string    `json:"sheetsLink"`
	ManagerID  uuid.UUID `json:"managerID"`

	// TotalHours float64   `json:"totalHours"`
	// CreatedAt  time.Time `json:"createdAt"`
	// UpdatedAt  time.Time `json:"updatedAt"`
}

// type Project struct {
// 	ID       string `json:"id" db:"project_id" example:"1114675b-93d2-4d67-ad0c-8851b6134af2"`
// 	Name     string `json:"name" db:"name" example:"Behance"`
// 	Icon     string `json:"icon" db:"icon" example:"https://prod-files-secure.s3.us-west-2.amazonaws.com/9a2e0635-b9d4-4178-a529-cf6b3bdce29d/7d460da2-42b7-4d5b-8d31-97a327675bc4/behance-1.svg?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Content-Sha256=UNSIGNED-PAYLOAD&X-Amz-Credential=AKIAT73L2G45HZZMZUHI%2F20241014%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20241014T055949Z&X-Amz-Expires=3600&X-Amz-Signature=c67998b0c68723e6efb6268baf917f6ae9e4902238a2b146cb054a6cda51c7cf&X-Amz-SignedHeaders=host&x-id=GetObject"`
// 	IconType string `json:"iconType" db:"icon_type" example:"file"`
// 	Status   string `json:"status" db:"status" example:"В работе"`
// 	Type     string `json:"type" db:"type" example:"Личный"`
// 	Manager  string `json:"manager" db:"manager" example:"Mark"`

// 	SheetsLink string `json:"sheetsLink" db:"sheets_link"`

// 	ManagerID string `json:"managerID" db:"manager_id"`

// 	TotalHours       float64 `json:"-" db:"total_hours"`
// 	ManagementTaskID string  `json:"-" db:"management_task_id"`
// 	TestingTaskID    string  `json:"-" db:"testing_task_id"`
// }
