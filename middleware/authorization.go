package middleware

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/tsawlen/matchingAppChatService/common/security"
)

func Auth(jwtToken string, userId int) (bool, error) {

	authorization := strings.TrimPrefix(jwtToken, "Bearer ")

	token, errToken := jwt.Parse(authorization, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		key, err := security.GetPublicToken()
		if err != nil {
			return nil, err
		}
		return key, nil
	})
	if errToken != nil {
		_, ok := errToken.(*jwt.ValidationError)
		if ok {
			return false, errors.New("Unauthorized!")
		}
		return false, errToken
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			return false, errors.New("Unauthorized!")
		}
		if float64(userId) != claims["user"].(float64) {
			return false, errors.New("Unauthorized!")
		}
		return true, nil
	} else {
		return false, errors.New("Unauthorized!")
	}

}
