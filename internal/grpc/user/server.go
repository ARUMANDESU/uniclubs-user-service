package user

import (
	userv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/user"
	"google.golang.org/grpc"
)

type serverApi struct {
	userv1.UnimplementedUserServer
	auth       Auth
	management Management
}

func Register(gRPC *grpc.Server, auth Auth, management Management) {
	userv1.RegisterUserServer(gRPC, &serverApi{auth: auth, management: management})
}
