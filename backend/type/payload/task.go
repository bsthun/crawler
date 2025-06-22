package payload

type TaskSubmitRequest struct {
	Category *string `json:"category" validate:"required"`
	Type     *string `json:"type" validate:"required,oneof=web doc youtube"`
	Url      *string `json:"url" validate:"required,url"`
}

type TaskSubmitResponse struct {
	TaskId *uint64 `json:"taskId"`
}
