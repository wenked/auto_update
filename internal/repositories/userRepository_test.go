package repository

import (
	"auto-update/internal/database/models"
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUseryRepository(db)
	companyRepo := NewCompanyRepository(db)

	company := models.Company{Name: "Test company"}

	companyId, err := companyRepo.CreateCompany(context.Background(), company)
	assert.NoError(t, err)

	user := models.User{
		Name:      "Test",
		Email:     "test@test.com",
		Password:  "123456",
		CompanyID: companyId,
	}

	id, err := repo.CreateUser(context.Background(), user)

	assert.NoError(t, err)
	assert.NotEqual(t, 0, id)
}

func TestUpdateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUseryRepository(db)
	companyRepo := NewCompanyRepository(db)

	company := models.Company{Name: "Test company"}

	companyId, err := companyRepo.CreateCompany(context.Background(), company)
	assert.NoError(t, err)

	user := models.User{
		Name:      "Test",
		Email:     "test@test.com",
		Password:  "123456",
		CompanyID: companyId,
	}

	id, err := repo.CreateUser(context.Background(), user)

	assert.NoError(t, err)

	updateUser := models.User{ID: id, Name: "Update Test"}

	err = repo.UpdateUser(context.Background(), updateUser)
	assert.NoError(t, err)

	var name string

	err = db.QueryRow(`SELECT name FROM users where id = ?`, id).Scan(&name)
	assert.NoError(t, err)
	assert.Equal(t, updateUser.Name, name)
}
