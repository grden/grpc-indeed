package gapi

import (
	"context"

	"github.com/grden/indeed/server/auth"
	pb "github.com/grden/indeed/server/pb"

	"github.com/grden/indeed/server/token"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) Login(ctx context.Context, req *pb.LoginRequestMessage) (*pb.LoginResponseMessage, error) {
	user, err := auth.LoginUser(req.GetEmail(), req.GetPassword(), server.dbCollection.Users, context.TODO())
	if err != nil {
		if err == auth.ErrUserNotFound {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		if err == auth.ErrInvalidCredentials {
			return nil, status.Errorf(codes.Unauthenticated, err.Error())
		}
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	//

	resp := auth.ConvertUserObjectToUser(user)

	token, err := auth.CreateToken(server.tokenMaker, resp.Email, int64(resp.Id), server.config.AccessTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &pb.LoginResponseMessage{
		User:        resp,
		AccessToken: token,
	}, nil
}

func (server *Server) SignUp(ctx context.Context, req *pb.SignupRequestMessage) (*pb.SignupResponseMessage, error) {

	user, err := auth.RegisterUser(req.GetEmail(), req.GetPassword(), req.GetName(), server.dbCollection.Users, context.TODO())
	if err != nil {
		if err == auth.ErrUserAlreadyRegistered {
			return nil, status.Errorf(codes.AlreadyExists, err.Error())
		}
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	resp := auth.ConvertUserObjectToUser(user)

	return &pb.SignupResponseMessage{
		User: resp,
	}, nil
}
func (server *Server) GetUser(ctx context.Context, req *pb.EmptyRequest) (*pb.GetUserResponse, error) {
	payload, ok := ctx.Value(payloadHeader).(*token.Payload)
	if !ok {

		return nil, status.Errorf(codes.Internal, "missing required token")
	}

	user, err := auth.GetUser(server.dbCollection.Users, context.TODO(), payload.Email)

	if err != nil {
		if err == auth.ErrUserNotFound {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &pb.GetUserResponse{
		User: auth.ConvertUserObjectToUser(user),
	}, nil
}

func (server *Server) GetUsers(ctx context.Context, req *pb.UsersListRequest) (*pb.ListUserMessage, error) {
	payload, ok := ctx.Value(payloadHeader).(*token.Payload)
	if !ok {
		return nil, status.Errorf(codes.Internal, "missing required token")
	}
	users, err := auth.GetUsers(server.dbCollection.Users, context.TODO(), int(req.GetPageNumber()), int(req.GetPageSize()), req.Name, payload.Email)

	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	pbUsers := []*pb.User{}

	for _, user := range users {
		pbUser := auth.ConvertUserObjectToUser(user)
		pbUsers = append(pbUsers, pbUser)
	}

	return &pb.ListUserMessage{
		TotalCount: int32(len(pbUsers)),
		Users:      pbUsers,
	}, nil
}
