package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	accountv1 "github.com/TSX97/INK3/proto/account/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RegistrationGateway struct {
	client accountv1.AccountServiceClient
}

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "registration")

	slog.SetDefault(logger)
	conn, _ := grpc.Dial("account:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer conn.Close()

	slog.Info("connect",
		"protocol", "gRPC",
		"port", ":50051",
	)

	gw := &RegistrationGateway{
		client: accountv1.NewAccountServiceClient(conn),
	}

	http.HandleFunc("/register", gw.handleRegister)

	slog.Info("start serve",
		"protocol", "http",
		"port", ":8081",
	)
	http.ListenAndServe(":8081", nil)
}

func (gw *RegistrationGateway) handleRegister(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		slog.Error("HTTP method not allowed",
			"method", r.Method,
			"expected", http.MethodPost,
		)

		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var protoReq accountv1.RegUserRequest
	if err := json.NewDecoder(r.Body).Decode(&protoReq); err != nil {
		slog.Warn("invalid json", "err", err)
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	slog.Info("register request",
		"nickname", protoReq.GetName(),
		"email", protoReq.GetEmail(),
	)

	resp, err := gw.client.Register(r.Context(), &protoReq)
	slog.Info("received a response",
		"response", resp,
		"protocol", "gRPC",
	)
	if err != nil {
		slog.Warn("grpc Register failed", "err", err, "nickname", protoReq.GetName())
		http.Error(w, "account service unavailable", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("failed to encode response", "err", err)
	}

}
