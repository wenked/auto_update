package server

import (
	"auto-update/internal/database/models"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func compareHashPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

type jwtCustomClaims struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	jwt.RegisteredClaims
}

func (s *Server) loginHandler(c echo.Context) error {
	session := new(models.LoginRequest)

	if err := c.Bind(session); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	user, err := s.db.GetUserByEmail(session.Email)

	if err != nil {
		slog.Error("error getting user", err)
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
	}

	if !compareHashPassword(session.Password, user.Password) {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid email or password"})
	}

	claims := &jwtCustomClaims{
		user.ID,
		user.Name,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		slog.Error("error signing token", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"ID":    user.ID,
		"Name":  user.Name,
		"Email": user.Email,
		"token": t,
	})

}
