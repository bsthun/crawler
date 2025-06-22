package payload

import "time"

type TaskSubmitRequest struct {
	Category *string `json:"category" validate:"required"`
	Type     *string `json:"type" validate:"required,oneof=web doc youtube"`
	Url      *string `json:"url" validate:"required,url"`
}

type TaskSubmitResponse struct {
	TaskId *uint64 `json:"taskId"`
}

type TaskListRequest struct {
	UploadId *uint64 `json:"uploadId"`
	Limit    *int32  `json:"limit"`
	Offset   *int32  `json:"offset"`
}

type TaskListItem struct {
	Id           *uint64    `json:"id"`
	UserId       *uint64    `json:"userId"`
	UploadId     *uint64    `json:"uploadId"`
	CategoryId   *uint64    `json:"categoryId"`
	Type         *string    `json:"type"`
	Url          *string    `json:"url"`
	Status       *string    `json:"status"`
	FailedReason *string    `json:"failedReason"`
	TokenCount   *int32     `json:"tokenCount"`
	CreatedAt    *time.Time `json:"createdAt"`
	UpdatedAt    *time.Time `json:"updatedAt"`
}

type TaskListResponse struct {
	Count *uint64         `json:"count"`
	Tasks []*TaskListItem `json:"tasks"`
}
