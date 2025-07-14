package service

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/pkg"
	"cmf/paint_proj/repository"
	"context"
)

type UserService interface {
	LoginHandler(ctx context.Context, req *model.LoginRequest) (int64, string, error)
	UpdateUserInfo(ctx context.Context, userId int64, req *model.UpdateUserInfoRequest) error
}
type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(pr repository.UserRepository) UserService {
	return &userService{
		userRepo: pr,
	}
}

func (u userService) LoginHandler(ctx context.Context, req *model.LoginRequest) (int64, string, error) {
	// 1.首次登录小程序时，通过 code换取 openid 的一步 <这一步只需要在用户第一次进入时调用即可>
	// 之后应该缓存用户的 openid（或你生成的 userID），在用户每次发请求时带上，后端只需要解析 token 并还原 userID
	openid, err := pkg.GetOpenIDByCode(req.Code)
	if err != nil {
		return 0, "", err
	}
	// 2.获取或创建用户
	user, err := u.userRepo.GetOrCreateUserByOpenID(openid, req.Nickname, req.Avatar)
	if err != nil {
		return 0, "", err
	}
	// 3.生成自定义 token（推荐 JWT）
	token, _ := pkg.GenerateJWTToken(user.ID)
	return user.ID, token, nil
}

func (u userService) UpdateUserInfo(ctx context.Context, userId int64, req *model.UpdateUserInfoRequest) error {
	return u.userRepo.UpdateUserInfo(userId, req)
}
