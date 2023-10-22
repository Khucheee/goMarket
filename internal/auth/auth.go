package auth

import (
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"time"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

const TokenExp = 90 * time.Minute
const SecretKey = "supersecretkey" //надо будет хранить эту штуку в базе в идеале

// тут создаем куку, которую будем выдавать
func MakeAuthCookie(userID string) (*http.Cookie, error) {
	//собираем JWT токен
	tokenString, err := makeToken(userID)
	if err != nil {
		return nil, err
	}
	//устанавливаем заголовки в куки
	newCookie := &http.Cookie{
		Name:  "auth",
		Value: tokenString,
	}
	return newCookie, nil
}

// собираем JWT токен
func makeToken(userID string) (string, error) {
	//создаем jwt токен передавая метод шифрования, полезную нагрузку(время действия токена и userID)
	jt := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: userID,
	})
	//подписываем его своим ключом
	tokenString, err := jt.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil //возвращаем токен
}

func ParseToken(tokenString string) (string, error) { //парсим токен
	//создаем структуру в которую будем парсить токен
	claims := &Claims{}
	jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	return claims.UserID, nil
}

func ParseUserFromCookie(r *http.Request) (string, error) {
	c, err := r.Cookie("auth")
	if err != nil {
		return "", err
	}
	uid, err := ParseToken(c.Value)
	if err != nil {
		return "", err
	}
	return uid, nil
}
