package payload

type StateResponse struct {
	UserId      *uint64 `json:"userId"`
	DisplayName *string `json:"displayName"`
	Email       *string `json:"email"`
	PhotoUrl    *string `json:"photoUrl"`
	IsAdmin     *bool   `json:"isAdmin"`
}
