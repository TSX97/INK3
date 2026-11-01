package main

import (
	"context"
	"fmt"
	"net"
	"os"

	"log/slog"

	"github.com/TSX97/INK3/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	accountv1 "github.com/TSX97/INK3/proto/account/v1"
)

type User struct {
	Id       int
	Name     string
	Email    string
	Password string
}

type UserRepository interface {
	CreateUser(ctx context.Context, user User) (int, error)
}

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context, user User) (int, error) {
	var id int
	err := r.pool.QueryRow(ctx, "INSERT INTO Users (nickname, email, password) VALUES ($1, $2, $3) RETURNING id", user.Name, user.Email, user.Password).Scan(&id)
	if err != nil {
		slog.Error("database query execution failed", "query", "INSERT INTO Users", "error", err)
		return 0, err
	}
	return id, nil
}

type AccountService struct {
	users UserRepository
}

func (s *AccountService) Register(ctx context.Context, user User) (int, error) {
	return s.users.CreateUser(ctx, user)
}

type AccountServiceServer struct {
	accountv1.UnimplementedAccountServiceServer

	service *AccountService
}

func (ass *AccountServiceServer) Register(ctx context.Context, req *accountv1.RegUserRequest) (*accountv1.RegUserResponse, error) {
	name := req.GetName()
	email := req.GetEmail()
	password := req.GetPassword()

	user := User{Name: name, Email: email, Password: password}
	id, err := ass.service.Register(ctx, user)
	if err != nil {
		slog.Info("failed to register user", "err", err, "name", name)
		return nil, err
	}

	slog.Info("user registered successfully", "user_id", id)
	return &accountv1.RegUserResponse{
		Id: int32(id),
	}, nil
}

func (ass *AccountServiceServer) AuthUser(ctx context.Context, req *accountv1.AuthUserRequest) (*accountv1.AuthUserResponse, error) {
	// Временная gRPC заглушка, чтобы прошел коммит
	return &accountv1.AuthUserResponse{Id: 0}, nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "account")
	slog.SetDefault(logger)

	pool, err := db.NewPool()
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	repo := &PostgresUserRepository{pool: pool}
	accService := &AccountService{users: repo}
	server := &AccountServiceServer{service: accService}

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}

	grpcServer := grpc.NewServer()
	accountv1.RegisterAccountServiceServer(grpcServer, server)

	slog.Info("start serve on localhost:50051")
	if err := grpcServer.Serve(listener); err != nil {
		panic(fmt.Sprintf("failed to listen: %v", err))
	}
}

