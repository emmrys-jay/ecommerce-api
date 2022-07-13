package util

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func RandomString() string {
	text := ""
	for i := 0; i < 12; i++ {
		char := string(rand.Intn(26) + 97)
		text += char
	}
	return text
}
