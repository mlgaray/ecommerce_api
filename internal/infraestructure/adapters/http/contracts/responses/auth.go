package responses

// AuthResponse represents auth data in responses.
type AuthResponse struct {
	Token string `json:"token"`
}

// SignInResponse represents the successful sign in response.
type SignInResponse struct {
	Auth AuthResponse `json:"auth"`
}
