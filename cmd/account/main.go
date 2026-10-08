package main

import (
	"context"
	"fmt"

	"github.com/TSX97/INK3/db"
	"github.com/jackc/pgx/v5/pgxpool"

	accountv1 "github.com/TSX97/INK3/proto/account/v1"
)

type User struct {
	Id       int
	Name     string
	Email    string
	Password string
}

type UserRepository interface {
	CreateUser(ctx context.Context, user User) (bool, error)
	GetAllUsers(ctx context.Context) ([]User, error)
}

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context, user User) (bool, error) {
	_, err := r.pool.Exec(ctx, "INSERT INTO Users (id, name, email, password) VALUES ($1, $2, $3)", user.Id, user.Name, user.Email, user.Password)
	if err != nil {
		return false, err
	}
	return true, err
}

func (r *PostgresUserRepository) GetAllUsers(ctx context.Context) ([]User, error) {
	rows, err := r.pool.Query(ctx, "SELECT * FROM Users")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []User

	for rows.Next() {
		var user User

		err := rows.Scan(&user.Id, &user.Name, &user.Email, &user.Password)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

type AccountService struct {
	users UserRepository
}

func (s *AccountService) Register(ctx context.Context, user User) (bool, error) {
	return s.users.CreateUser(ctx, user)
}

type AccountServiceServer struct {
	accountv1.UnimplementedAccountServiceServer

	service *AccountService
}

func (ass *AccountServiceServer) Register(ctx context.Context, req *accountv1.RegUserRequest) (*accountv1.RegUserResponse, error) {
	id := req.GetId()
	name := req.GetName()
	email := req.GetEmail()
	password := req.GetPassword()

	user := User{Id: int(id), Name: name, Email: email, Password: password}
	flag, err := ass.service.Register(ctx, user)
	if err != nil {
		return nil, err
	}

	if flag == true {
		return &accountv1.RegUserResponse{
			Id:      id,
			Success: true,
			Error:   accountv1.RegError_ERROR_UNDEFINED,
		}, nil

	} else {
		return &accountv1.RegUserResponse{
			Id:      0,
			Success: false,
			Error:   accountv1.RegError_ERROR_UNDEFINED,
		}, nil

	}

}

func main() {

	pool, err := db.NewPool()
	if err != nil {
		panic(err)
	}

	defer pool.Close()

	err = db.Ping(context.Background(), pool)
	if err != nil {
		panic(err)
	}

	fmt.Println("start serve on localhost:8080")
}
