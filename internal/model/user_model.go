package model

type User struct {
	ID             int64
	Login          string
	PasswordHash   string
	CurrentBalance float64
	WithdrawnTotal float64
}

// DTO
type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}
