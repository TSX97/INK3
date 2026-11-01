package main

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	accountv1 "github.com/TSX97/INK3/proto/account/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

type AuthGateway struct {
	client accountv1.AccountServiceClient
}

func main() {
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	).With("service", "authentication")

	slog.SetDefault(logger)

	const (
		accountAddress = "account:50051"
		httpAddress    = ":8082"
	)

	conn, err := grpc.Dial(
		accountAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		slog.Error("failed to connect to account service",
			"address", accountAddress,
			"error", err,
		)
		os.Exit(1)
	}
	defer conn.Close()

	slog.Info("connected to account service",
		"protocol", "grpc",
		"address", accountAddress,
	)

	gw := &AuthGateway{
		client: accountv1.NewAccountServiceClient(conn),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/auth", gw.handleAuth)

	slog.Info("authentication service started",
		"protocol", "http",
		"address", httpAddress,
	)

	if err := http.ListenAndServe(httpAddress, mux); err != nil {
		slog.Error("http server stopped",
			"address", httpAddress,
			"error", err,
		)
		os.Exit(1)
	}
}

func (gw *AuthGateway) handleAuth(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	logger := slog.With(
		"method", r.Method,
		"path", r.URL.Path,
		"remote_addr", r.RemoteAddr,
	)

	statusCode := http.StatusOK

	defer func() {
		logger.Info("request completed",
			"status", statusCode,
			"duration", time.Since(start).String(),
		)
	}()

	respond := func(code int, message string) {
		statusCode = code
		http.Error(w, message, code)
	}

	if r.Method != http.MethodPost {
		logger.Warn("method not allowed")
		respond(http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	defer r.Body.Close()

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		logger.Warn("failed to read request body", "error", err)
		respond(http.StatusBadRequest, "invalid request body")
		return
	}

	var protoReq accountv1.AuthUserRequest
	if err := protojson.Unmarshal(body, &protoReq); err != nil {
		logger.Warn("invalid JSON", "error", err)
		respond(http.StatusBadRequest, "invalid JSON")
		return
	}

	logger = logger.With("nickname", protoReq.GetName())

	if protoReq.GetName() == "" || protoReq.GetPassword() == "" {
		logger.Warn("missing authentication credentials")
		respond(http.StatusBadRequest, "name and password are required")
		return
	}

	resp, err := gw.client.AuthUser(r.Context(), &protoReq)
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			logger.Error("unexpected grpc error", "error", err)
			respond(http.StatusBadGateway, "account service unavailable")
			return
		}

		switch st.Code() {
		case codes.Unauthenticated:
			logger.Warn("authentication failed",
				"grpc_code", st.Code().String(),
			)
			respond(http.StatusUnauthorized, "invalid credentials")

		case codes.InvalidArgument:
			logger.Warn("invalid authentication request",
				"grpc_code", st.Code().String(),
			)
			respond(http.StatusBadRequest, "invalid request")

		case codes.Unavailable:
			logger.Error("account service unavailable",
				"grpc_code", st.Code().String(),
			)
			respond(http.StatusServiceUnavailable, "account service unavailable")

		default:
			logger.Error("authentication service error",
				"grpc_code", st.Code().String(),
				"grpc_message", st.Message(),
			)
			respond(http.StatusBadGateway, "authentication failed")
		}
		return
	}

	responseBody, err := protojson.Marshal(resp)
	if err != nil {
		logger.Error("failed to encode response", "error", err)
		respond(http.StatusInternalServerError, "internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(responseBody); err != nil {
		logger.Error("failed to write response", "error", err)
		return
	}

	logger.Info("authentication successful",
		"user_id", resp.GetId(),
	)
}
