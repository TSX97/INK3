package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/TSX97/INK3/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	accountv1 "github.com/TSX97/INK3/proto/account/v1"
)

type User struct {
	Id         int
	Name       string
	Email      string
	Created_at time.Time
	Password   string
}

type UserRepository interface {
	CreateUser(ctx context.Context, user User) (bool, error)
	GetAllUsers(ctx context.Context) ([]User, error)
}

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context, user User) (bool, error) {
	_, err := r.pool.Exec(ctx, "INSERT INTO Users (id, name, email, password) VALUES ($1, $2, $3, $4)", user.Id, user.Name, user.Email, user.Password)
	if err != nil {
		return false, err
	}
	return true, err
}

func (r *PostgresUserRepository) GetAllUsers(ctx context.Context) ([]User, error) {
	rows, err := r.pool.Query(ctx, "SELECT id, name, email, created_at, password FROM Users")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []User

	for rows.Next() {
		var user User

		err := rows.Scan(&user.Id, &user.Name, &user.Email, &user.Created_at, &user.Password)
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

func (s *AccountService) GetAllUsers(ctx context.Context) ([]User, error) {
	return s.users.GetAllUsers(ctx)
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
			Error:   accountv1.Error_UNDEFINED,
		}, nil

	} else {
		return &accountv1.RegUserResponse{
			Id:      0,
			Success: false,
			Error:   accountv1.Error_UNDEFINED,
		}, nil

	}

}

func (ass *AccountServiceServer) GetAllUsers(ctx context.Context, req *accountv1.GetAllRequest) (*accountv1.GetAllResponse, error) {
	users, err := ass.service.GetAllUsers(ctx)
	if err != nil {
		return &accountv1.GetAllResponse{Success: false, Error: accountv1.Error_UNDEFINED}, nil
	}

	var protoUsers []*accountv1.UserInfo
	for _, u := range users {
		protoUsers = append(protoUsers, &accountv1.UserInfo{
			Id:        int32(u.Id),
			Name:      u.Name,
			Email:     u.Email,
			CreatedAt: timestamppb.New(u.Created_at),
		})
	}

	return &accountv1.GetAllResponse{
		Success: true,
		Result:  protoUsers,
		Error:   accountv1.Error_UNDEFINED,
	}, nil

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

	repo := &PostgresUserRepository{pool: pool}
	accService := &AccountService{users: repo}
	server := &AccountServiceServer{service: accService}

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		panic(fmt.Sprintf("failed to listen on port 50051: %v", err))
	}

	grpcServer := grpc.NewServer()
	accountv1.RegisterAccountServiceServer(grpcServer, server)

	fmt.Println("start serve on localhost:50051")
	if err := grpcServer.Serve(listener); err != nil {
		panic(fmt.Sprintf("failed to serve gRPC: %v", err))
	}

}
