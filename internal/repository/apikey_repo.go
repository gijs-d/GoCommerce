package repository

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"webshop/internal/database"
)

type ApiKey struct {
	ID                 string    `json:"id"`
	Description        string    `json:"description"`
	ConsumerKey        string    `json:"consumer_key"`
	ConsumerSecretHash string    `json:"-"`
	Permissions        string    `json:"permissions"` // 'read', 'write', 'read_write'
	CreatedAt          time.Time `json:"created_at"`
}

type ApiKeyRepo struct {
	db *database.DB
}

func NewApiKeyRepo(db *database.DB) *ApiKeyRepo {
	return &ApiKeyRepo{db: db}
}

func hashSecret(secret string) string {
	h := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(h[:])
}

func generateKey(prefix string, byteLength int) (string, error) {
	bytes := make([]byte, byteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(bytes), nil
}

// GenerateKey maakt een officiële WooCommerce consumer key en secret aan
func (r *ApiKeyRepo) GenerateKey(ctx context.Context, description, permissions string) (string, string, *ApiKey, error) {
	consumerKey, err := generateKey("ck_", 20)
	if err != nil {
		return "", "", nil, err
	}

	consumerSecret, err := generateKey("cs_", 20)
	if err != nil {
		return "", "", nil, err
	}

	secretHash := hashSecret(consumerSecret)

	if permissions != "read" && permissions != "write" && permissions != "read_write" {
		permissions = "read_write"
	}

	query := `
		INSERT INTO api_keys (description, consumer_key, consumer_secret_hash, permissions)
		VALUES ($1, $2, $3, $4)
		RETURNING id, description, consumer_key, permissions, created_at
	`
	key := &ApiKey{}
	err = r.db.Pool.QueryRow(ctx, query, description, consumerKey, secretHash, permissions).
		Scan(&key.ID, &key.Description, &key.ConsumerKey, &key.Permissions, &key.CreatedAt)
	if err != nil {
		return "", "", nil, err
	}

	return consumerKey, consumerSecret, key, nil
}

// ValidateKey verifieert of een consumer_key en secret overeenkomen en geldig zijn
func (r *ApiKeyRepo) ValidateKey(ctx context.Context, consumerKey, consumerSecret string) (string, error) {
	query := `
		SELECT consumer_secret_hash, permissions
		FROM api_keys
		WHERE consumer_key = $1
	`
	var storedHash, permissions string
	err := r.db.Pool.QueryRow(ctx, query, consumerKey).Scan(&storedHash, &permissions)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errors.New("ongeldige WooCommerce API-sleutel")
		}
		return "", err
	}

	providedHash := hashSecret(consumerSecret)
	if storedHash != providedHash {
		return "", errors.New("onjuist WooCommerce API-secret")
	}

	return permissions, nil
}

func (r *ApiKeyRepo) ListAll(ctx context.Context) ([]ApiKey, error) {
	query := `
		SELECT id, description, consumer_key, permissions, created_at
		FROM api_keys
		ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []ApiKey
	for rows.Next() {
		var k ApiKey
		if err := rows.Scan(&k.ID, &k.Description, &k.ConsumerKey, &k.Permissions, &k.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}

func (r *ApiKeyRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM api_keys WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}