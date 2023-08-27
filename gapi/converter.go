package gapi

import (
	"time"

	db "github.com/uwemakan/simplebank/db/sqlc"
	"github.com/uwemakan/simplebank/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ConvertUser(user db.User) *pb.User {
	return &pb.User{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: timestamppb.New(user.PasswordChangedAt),
		CreatedAt:         timestamppb.New(user.CreatedAt),
	}
}

func ConvertLoginResponse(
	user db.User, session db.Session,
	accessToken, refreshToken string,
	accessExpiredAt, refreshExpiredAt time.Time,
) *pb.LoginUserResponse {
	return &pb.LoginUserResponse{
		SessionId:             session.ID.String(),
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  timestamppb.New(accessExpiredAt),
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: timestamppb.New(refreshExpiredAt),
		User:                  ConvertUser(user),
	}
}
