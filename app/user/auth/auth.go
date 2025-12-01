package auth

import (
	"context"

	"github.com/byteflowing/base/app/user/dal/query"
	userv1 "github.com/byteflowing/proto/gen/go/user/v1"
)

type Auth interface {
	Authenticate(ctx context.Context, req *userv1.SignInReq, tx *query.Query) (*userv1.SignInResult, error)
}
