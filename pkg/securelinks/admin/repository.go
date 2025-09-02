package admin

import (
	"database/sql"
	"secure-email-mvp/pkg/models"
)

// Repository interface for admin operations
type Repository interface {
	GetAdminByEmail(email string) (*models.AdminUser, error)
	CreateAdmin(user *models.AdminUser) error
}

// SQLiteAdminRepository implements Repository for SQLite
type SQLiteAdminRepository struct {
	db *sql.DB
}

// NewSQLiteAdminRepository creates a new SQLite admin repository
func NewSQLiteAdminRepository(db *sql.DB) *SQLiteAdminRepository {
	return &SQLiteAdminRepository{db: db}
}

// GetAdminByEmail retrieves an admin user by email
func (r *SQLiteAdminRepository) GetAdminByEmail(email string) (*models.AdminUser, error) {
	query := `
		SELECT id, email, password, role, created_at
		FROM admin_users
		WHERE email = ? AND role = 'admin'
	`
	
	var user models.AdminUser
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return &user, nil
}

// CreateAdmin creates a new admin user
func (r *SQLiteAdminRepository) CreateAdmin(user *models.AdminUser) error {
	query := `
		INSERT INTO admin_users (email, password, role, created_at)
		VALUES (?, ?, ?, ?)
	`
	
	result, err := r.db.Exec(query, user.Email, user.Password, user.Role, user.CreatedAt)
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	user.ID = id
	return nil
}







