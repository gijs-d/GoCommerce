package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"webshop/internal/auth"
	"webshop/internal/database"
)

type UserRepo struct {
	db *database.DB
}

func NewUserRepo(db *database.DB) *UserRepo {
	return &UserRepo{db: db}
}

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (r *UserRepo) CreateUser(ctx context.Context, email, passwordHash, fullName, role string) (*User, error) {
	query := `
		INSERT INTO users (email, password_hash, full_name, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, full_name, role, created_at, updated_at
	`
	u := &User{}
	err := r.db.Pool.QueryRow(ctx, query, email, passwordHash, fullName, role).
		Scan(&u.ID, &u.Email, &u.FullName, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, password_hash, full_name, role, created_at, updated_at
		FROM users
		WHERE LOWER(email) = LOWER($1)
	`
	u := &User{}
	err := r.db.Pool.QueryRow(ctx, query, email).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, userID string) (*User, error) {
	query := `
		SELECT id, email, password_hash, full_name, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	u := &User{}
	err := r.db.Pool.QueryRow(ctx, query, userID).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

// CreateSession maakt een sessie aan in de database
func (r *UserRepo) CreateSession(ctx context.Context, userID string) (string, error) {
	token, err := auth.GenerateSessionToken()
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(auth.SessionDuration())
	query := `
		INSERT INTO sessions (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err = r.db.Pool.Exec(ctx, query, userID, token, expiresAt)
	if err != nil {
		return "", err
	}
	return token, nil
}

// GetUserBySession haalt de gebruiker op indien sessie geldig en nog niet verlopen is
func (r *UserRepo) GetUserBySession(ctx context.Context, token string) (*auth.SessionUser, error) {
	query := `
		SELECT u.id, u.email, u.full_name, u.role
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token = $1 AND s.expires_at > now()
	`
	u := &auth.SessionUser{}
	err := r.db.Pool.QueryRow(ctx, query, token).Scan(&u.ID, &u.Email, &u.FullName, &u.Role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

// DeleteSession verwijdert de sessie (uitloggen)
func (r *UserRepo) DeleteSession(ctx context.Context, token string) error {
	query := `DELETE FROM sessions WHERE token = $1`
	_, err := r.db.Pool.Exec(ctx, query, token)
	return err
}

// DeleteUserSessions verwijdert alle sessies van een gebruiker (bij bijv. wachtwoordreset of force logout)
func (r *UserRepo) DeleteUserSessions(ctx context.Context, userID string) error {
	query := `DELETE FROM sessions WHERE user_id = $1`
	_, err := r.db.Pool.Exec(ctx, query, userID)
	return err
}

type CustomerStats struct {
	ID           string     `json:"id"`
	FullName     string     `json:"full_name"`
	Email        string     `json:"email"`
	TotalOrders  int        `json:"total_orders"`
	TotalSpent   float64    `json:"total_spent"`
	RegisteredAt time.Time  `json:"registered_at"`
	LastOrderAt  *time.Time `json:"last_order_at,omitempty"`
}

// ListCustomersWithStats haalt klanten op met hun totale bestedingen en bestelaantallen
func (r *UserRepo) ListCustomersWithStats(ctx context.Context) ([]CustomerStats, error) {
	query := `
		SELECT 
			u.id, u.full_name, u.email,
			COUNT(o.id) as total_orders,
			COALESCE(SUM(o.total_amount), 0) as total_spent,
			u.created_at as registered_at,
			MAX(o.created_at) as last_order_at
		FROM users u
		LEFT JOIN orders o ON o.user_id = u.id AND o.payment_status = 'paid'
		WHERE u.role = 'customer'
		GROUP BY u.id
		ORDER BY total_spent DESC, u.created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customers []CustomerStats
	for rows.Next() {
		var c CustomerStats
		if err := rows.Scan(&c.ID, &c.FullName, &c.Email, &c.TotalOrders, &c.TotalSpent, &c.RegisteredAt, &c.LastOrderAt); err == nil {
			customers = append(customers, c)
		}
	}
	return customers, nil
}