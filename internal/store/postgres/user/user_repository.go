package users

import (
	"context"
	"os/user"

	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *userRepository {
	return &userRepository{
		db: db,
	}
}

func (u *userRepository) Get(ctx context.Context, email, password string) (*user.User, error) {
	query := `SELECT id, name, email, broker, photo_path FROM users WHERE email=$1 AND password=$2`

	var user user.User
	err := u.db.QueryRow(ctx, query, email, password).Scan(&user.ID, &user.Name, &user.Email, &user.Broker, &user.PhotoPath)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// func (u *userRepository) Create(email, password, name, broker string, photoPath *string) (*user.User, error) {
// 	query := `INSERT INTO users (name, email, password, broker, photo_path) VALUES ($1, $2, $3, $4, $5) RETURNING id`

// 	var id int64
// 	err := u.db.QueryRow(context.Background(), query, name, email, password, broker, photoPath).Scan(&id)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &user.User{
// 		ID:        id,
// 		Name:      name,
// 		Email:     email,
// 		Broker:    broker,
// 		PhotoPath: photoPath,
// 	}, nil
// }

// func (u *userRepository) Delete(email, password string) error {
// 	query := `DELETE FROM users WHERE email=$1 AND password=$2`
// 	_, err := u.db.Exec(context.Background(), query, email, password)
// 	return err
// }
