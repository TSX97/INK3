package main

import (
	"context"
	"encoding/json"
	"net/http"

	accountv1 "github.com/TSX97/INK3/proto/account/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RegistrationGateway struct {
	client accountv1.AccountServiceClient
}

func main() {
	conn, _ := grpc.Dial("account:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer conn.Close()

	gw := &RegistrationGateway{
		client: accountv1.NewAccountServiceClient(conn),
	}

	http.HandleFunc("/register", gw.handleRegister)
	http.ListenAndServe(":8081", nil)
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
