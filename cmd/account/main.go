package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/TSX97/INK3/db"
	accountv1 "github.com/TSX97/INK3/proto/account/v1"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type User struct {
	Id       int
	Name     string
	Email    string
	Password string
}

type UserRepository interface {
	CreateUser(ctx context.Context, user User) (int, error)
	AuthenticateUser(ctx context.Context, name, password string) (int, error)
}

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func (r *PostgresUserRepository) CreateUser(
	ctx context.Context,
	user User,
) (int, error) {
	var id int

	err := r.pool.QueryRow(
		ctx,
		`INSERT INTO Users (nickname, email, password)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		user.Name,
		user.Email,
		user.Password,
	).Scan(&id)

	if err != nil {
		slog.Error("database query execution failed",
			"operation", "create_user",
			"error", err,
		)
		return 0, err
	}

	return id, nil
}

func (r *PostgresUserRepository) AuthenticateUser(
	ctx context.Context,
	name, password string,
) (int, error) {
	var id int

	err := r.pool.QueryRow(
		ctx,
		`SELECT id FROM Users
		 WHERE nickname = $1 AND password = $2`,
		name,
		password,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

type AccountService struct {
	users UserRepository
}

func (s *AccountService) Register(
	ctx context.Context,
	user User,
) (int, error) {
	return s.users.CreateUser(ctx, user)
}

func (s *AccountService) Authenticate(
	ctx context.Context,
	name, password string,
) (int, error) {
	return s.users.AuthenticateUser(ctx, name, password)
}

type AccountServiceServer struct {
	accountv1.UnimplementedAccountServiceServer

	service *AccountService
}

func (ass *AccountServiceServer) Register(
	ctx context.Context,
	req *accountv1.RegUserRequest,
) (*accountv1.RegUserResponse, error) {
	user := User{
		Name:     req.GetName(),
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}

	id, err := ass.service.Register(ctx, user)
	if err != nil {
		slog.Error("failed to register user", "error", err)

		return nil, status.Error(
			codes.Internal,
			"failed to register user",
		)
	}

	slog.Info("user registered successfully", "user_id", id)

	return &accountv1.RegUserResponse{
		Id: int32(id),
	}, nil
}

func (ass *AccountServiceServer) AuthUser(
	ctx context.Context,
	req *accountv1.AuthUserRequest,
) (*accountv1.AuthUserResponse, error) {
	id, err := ass.service.Authenticate(
		ctx,
		req.GetName(),
		req.GetPassword(),
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(
				codes.Unauthenticated,
				"invalid credentials",
			)
		}

		slog.Error("failed to authenticate user", "error", err)

		return nil, status.Error(
			codes.Internal,
			"authentication failed",
		)
	}

	slog.Info("user authenticated successfully", "user_id", id)

	return &accountv1.AuthUserResponse{
		Id: int32(id),
	}, nil
}

func main() {
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, nil),
	).With("service", "account")

	slog.SetDefault(logger)

	pool, err := db.NewPool()
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	repo := &PostgresUserRepository{
		pool: pool,
	}

	accService := &AccountService{
		users: repo,
	}

	server := &AccountServiceServer{
		service: accService,
	}

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		panic(fmt.Sprintf("failed to listen: %v", err))
	}

	grpcServer := grpc.NewServer()

	accountv1.RegisterAccountServiceServer(
		grpcServer,
		server,
	)

	slog.Info("starting gRPC server", "address", ":50051")

	if err := grpcServer.Serve(listener); err != nil {
		panic(fmt.Sprintf("failed to serve: %v", err))
	}
}
