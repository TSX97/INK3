package main

import (
	"io"
	"log/slog"
	"net/http"
	"os"

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
		slog.NewJSONHandler(os.Stdout, nil),
	).With("service", "auth")

	slog.SetDefault(logger)

	conn, err := grpc.Dial(
		"account:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		slog.Error("failed to connect to account service", "err", err)
		os.Exit(1)
	}
	defer conn.Close()

	slog.Info("connect",
		"protocol", "gRPC",
		"address", "account:50051",
	)

	gw := &AuthGateway{
		client: accountv1.NewAccountServiceClient(conn),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/auth", gw.handleAuth)

	slog.Info("start serve",
		"protocol", "http",
		"port", ":8082",
	)

	if err := http.ListenAndServe(":8082", mux); err != nil {
		slog.Error("http server stopped", "err", err)
		os.Exit(1)
	}
}

func (gw *AuthGateway) handleAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	var protoReq accountv1.AuthUserRequest
	if err := protojson.Unmarshal(body, &protoReq); err != nil {
		slog.Warn("invalid JSON", "err", err)
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	slog.Info("auth request", "nickname", protoReq.GetName())

	resp, err := gw.client.AuthUser(r.Context(), &protoReq)
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			slog.Error("unexpected gRPC error", "err", err)
			http.Error(w, "account service unavailable", http.StatusBadGateway)
			return
		}

		slog.Warn("grpc AuthUser failed",
			"code", st.Code().String(),
			"nickname", protoReq.GetName(),
		)

		switch st.Code() {
		case codes.Unauthenticated:
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
		case codes.InvalidArgument:
			http.Error(w, "invalid request", http.StatusBadRequest)
		case codes.Unavailable:
			http.Error(w, "account service unavailable", http.StatusServiceUnavailable)
		default:
			http.Error(w, "authentication failed", http.StatusBadGateway)
		}
		return
	}

	responseBody, err := protojson.Marshal(resp)
	if err != nil {
		slog.Error("failed to encode response", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(responseBody); err != nil {
		slog.Error("failed to write response", "err", err)
	}
}
