package payload

import (
	"backend/type/common"
	"time"
)

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
	common.Paginate
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

type OverviewHistoryItem struct {
	Submitted *uint64 `json:"submitted"`
	Pending   *uint64 `json:"pending"`
	Completed *uint64 `json:"completed"`
	Failed    *uint64 `json:"failed"`
}

type Overview struct {
	Histories      []*OverviewHistoryItem `json:"histories"`
	TokenHistories *int32                 `json:"tokenHistories"`
	TokenCount     *int32                 `json:"tokenCount"`
	PoolTokenCount *int32                 `json:"poolTokenCount"`
}
