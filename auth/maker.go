package token

type Maker interface {
	CreateToken() string
	VerifyToken(token string) bool
}
