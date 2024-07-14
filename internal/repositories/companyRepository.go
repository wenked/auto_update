package repository

import (
	"auto-update/internal/database"
	"auto-update/internal/database/models"
	"database/sql"
	"errors"
	"log/slog"

	"golang.org/x/net/context"
)

type CompanyRepository interface {
	CreateCompany(ctx context.Context, company models.Company) (int64, error)
	UpdateCompany(ctx context.Context, company models.Company) error
	GetCompany(ctx context.Context, id int64) (models.Company, error)
	ListCompanies(ctx context.Context, page int64, limit int64) ([]models.Company, error)
	DeleteCompany(ctx context.Context, id int64) error
}

type SQLCompanyRepository struct {
	db *sql.DB
}

func NewCompanyRepository(db *sql.DB) *SQLCompanyRepository {
	return &SQLCompanyRepository{
		db: db,
	}
}

func (r *SQLCompanyRepository) CreateCompany(ctx context.Context, company models.Company) (int64, error) {
	query := "INSERT INTO companies (name) values(?)"
	result, err := r.db.ExecContext(ctx, query, company.Name)

	if err != nil {
		slog.Error("error creating company", "error", err)
		return 0, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		slog.Error("error getting id", "error", err)
		return 0, nil
	}

	return int64(id), nil
}

func (r *SQLCompanyRepository) UpdateCompany(ctx context.Context, company models.Company) error {
	query := "UPDATE companies SET name = ? WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, company.Name, company.ID)

	if err != nil {
		slog.Error("error updating company", "error", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		slog.Error("error getting rows affected", "error", err)
		return err
	}

	if rowsAffected == 0 {
		slog.Error("no rows affected", "error", err)
		return errors.New("no rows affected")
	}

	return nil
}

func (r *SQLCompanyRepository) GetCompany(ctx context.Context, id int64) (models.Company, error) {
	query := "SELECT * from companies WHERE id = ?"

	row := r.db.QueryRowContext(ctx, query, id)

	company, err := models.ScanRowCompany(row)

	if err != nil {
		slog.Error("error getting company", "error", err)
		return models.Company{}, err
	}

	return company, nil
}

func (r *SQLCompanyRepository) ListCompanies(ctx context.Context, page int64, limit int64) ([]models.Company, error) {
	offset := (page - 1) * limit

	query := "SELECT * FROM companies LIMIT ? OFFSET ?"

	rows, err := r.db.QueryContext(ctx, query, limit, offset)

	var companies []models.Company
	if err != nil {
		slog.Error("error in query", "error", err)
		return companies, err
	}

	defer rows.Close()

	companies, err = database.ScanRows(rows, models.ScanCompany)

	if err != nil {
		slog.Error("error getting comanies rows", "error", err)
		return companies, err
	}

	return companies, nil

}

func (r *SQLCompanyRepository) DeleteCompany(ctx context.Context, id int64) error {
	query := "DELETE FROM companies WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, id)

	if err != nil {
		slog.Error("error deleting company", "error", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("no rows affected")
	}
	return nil
}
