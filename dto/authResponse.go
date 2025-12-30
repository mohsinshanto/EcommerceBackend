package dto

type UserResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type RegisterResponse struct {
	Message string       `json:"message"`
	User    UserResponse `json:"user"`
}
