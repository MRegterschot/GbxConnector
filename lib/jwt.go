package lib

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/MRegterschot/GbxConnector/structs"
	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// Generates a JWT token.
func GenerateJWT(user structs.User) (string, error) {
	claims := jwt.MapClaims{
		"user": user,
		"exp":  time.Now().Add(time.Minute * 30).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		zap.L().Error("Failed to generate JWT token.", zap.Error(err))
		return "", err
	}

	return tokenString, nil
}

// Parses a JWT token.
func parseJWT(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			zap.L().Error("Invalid signing method.")
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret, nil
	})

	if err != nil {
		zap.L().Error("Failed to parse JWT token.", zap.Error(err))
		return nil, err
	}

	return token, nil
}

// Helper function to extract the token from "Bearer <token>"
func ExtractBearerToken(authHeader string) string {
	const prefix = "Bearer "
	if len(authHeader) > len(prefix) && authHeader[:len(prefix)] == prefix {
		return authHeader[len(prefix):]
	}
	return ""
}

// Validates a JWT token and returns the user.

func ValidateAndGetUser(tokenString string) (structs.User, error) {
	token, err := parseJWT(tokenString)
	if err != nil {
		return structs.User{}, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		zap.L().Error("Invalid JWT token.")
		return structs.User{}, jwt.ErrSignatureInvalid
	}

	userInterface, ok := claims["user"]
	if !ok {
		zap.L().Error("Failed to extract user from JWT claims: key 'user' not found.")
		return structs.User{}, errors.New("invalid token: user claim not found")
	}

	userMap, ok := userInterface.(map[string]interface{})
	if !ok {
		zap.L().Error("User claim is not a valid map.")
		return structs.User{}, errors.New("invalid token: user claim is not a valid map")
	}

	// Marshal back to JSON and then unmarshal into User struct
	userBytes, err := json.Marshal(userMap)
	if err != nil {
		zap.L().Error("Failed to marshal user claim to JSON.", zap.Error(err))
		return structs.User{}, err
	}

	var user structs.User
	err = json.Unmarshal(userBytes, &user)
	if err != nil {
		zap.L().Error("Failed to unmarshal user JSON into struct.", zap.Error(err))
		return structs.User{}, err
	}

	return user, nil
}
