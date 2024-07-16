package repository

import (
	"auto-update/internal/database/models"

	"golang.org/x/net/context"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user models.User) (int64, error)
}
  