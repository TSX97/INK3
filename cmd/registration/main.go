package main

import (
	"context"
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
	http.ListenAndServe(":8081", nil)

	slog.Info("start serve",
		"protocol", "http",
		"port", ":8081",
	)
}

func (gw *RegistrationGateway) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}

	var protoReq accountv1.RegUserRequest
	_ = json.NewDecoder(r.Body).Decode(&protoReq)

	resp, _ := gw.client.Register(context.Background(), &protoReq)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
