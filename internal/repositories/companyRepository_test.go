package repository

import (
	"auto-update/internal/database/models"
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory SQLite database: %v", err)
	}

	createTableQuery := `
        CREATE TABLE companies (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
    		name TEXT,
    		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
   			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
    `
	_, err = db.Exec(createTableQuery)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	return db
}

func TestCreateCompany(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewCompanyRepository(db)

	company := models.Company{Name: "Test Company"}

	id, err := repo.CreateCompany(context.Background(), company)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, id)

	var name string
	err = db.QueryRow("SELECT name FROM companies WHERE id = ?", id).Scan(&name)
	assert.NoError(t, err)
	assert.Equal(t, company.Name, name)
}

func TestUpdateCompany(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewCompanyRepository(db)
	company := models.Company{Name: "Test Company"}
	id, err := repo.CreateCompany(context.Background(), company)
	assert.NoError(t, err)

	updatedCompany := models.Company{ID: id, Name: "Updated Company"}
	err = repo.UpdateCompany(context.Background(), updatedCompany)
	assert.NoError(t, err)

	var name string
	err = db.QueryRow("SELECT name FROM companies WHERE id = ?", id).Scan(&name)
	assert.NoError(t, err)
	assert.Equal(t, updatedCompany.Name, name)
}

func TestGetCompany(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewCompanyRepository(db)
	company := models.Company{Name: "Test Company"}
	id, err := repo.CreateCompany(context.Background(), company)
	assert.NoError(t, err)

	result, err := repo.GetCompany(context.Background(), id)
	assert.NoError(t, err)
	assert.Equal(t, id, result.ID)
	assert.Equal(t, company.Name, result.Name)
}

func TestListCompanies(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewCompanyRepository(db)

	companies := []models.Company{
		{Name: "Company 1"},
		{Name: "Company 2"},
	}

	for _, company := range companies {
		_, err := repo.CreateCompany(context.Background(), company)
		assert.NoError(t, err)
	}

	result, err := repo.ListCompanies(context.Background(), 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, companies[0].Name, result[0].Name)
	assert.Equal(t, companies[1].Name, result[1].Name)
}

func TestDeleteCompany(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewCompanyRepository(db)

	company := models.Company{Name: "Test Company"}
	id, err := repo.CreateCompany(context.Background(), company)
	assert.NoError(t, err)

	err = repo.DeleteCompany(context.Background(), id)
	assert.NoError(t, err)

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM companies WHERE id = ?", id).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}
