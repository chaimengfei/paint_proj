package service

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/pkg"
	"cmf/paint_proj/repository"
	"context"
	"errors"

	"gorm.io/gorm"
)

type UserService interface {
	LoginHandler(ctx context.Context, req *model.LoginRequest) (int64, string, error)
	UpdateUserInfo(ctx context.Context, userId int64, req *model.UpdateUserInfoRequest) error

	// 后台用户管理
	CreateUserByAdmin(req *model.AdminUserAddRequest) (*model.User, error)
	GetUserByID(userID int64) (*model.User, error)
	UpdateUserByAdmin(req *model.AdminUserEditRequest) error
	GetUserList(page, pageSize int, keyword string) ([]*model.User, int64, error)
	DeleteUser(userID int64) error

	// 小程序用户绑定
	BindWechatToUser(userID int64, openid, wechatDisplayName string) error
	GetUserByMobilePhone(mobilePhone string) (*model.User, error)
	WechatBindMobile(userID int64, req *model.WechatBindMobileRequest) (*model.User, error)
}
type userService struct {
	userRepo repository.UserRepository
	shopRepo repository.ShopRepository
}

func NewUserService(pr repository.UserRepository, sr repository.ShopRepository) UserService {
	return &userService{
		userRepo: pr,
		shopRepo: sr,
	}
}

func (u userService) LoginHandler(ctx context.Context, req *model.LoginRequest) (int64, string, error) {
	// 1.首次登录小程序时，通过 code换取 openid 的一步 <这一步只需要在用户第一次进入时调用即可>
	// 之后应该缓存用户的 openid（或你生成的 userID），在用户每次发请求时带上，后端只需要解析 token 并还原 userID
	openid, err := pkg.GetOpenIDByCode(req.Code)
	if err != nil {
		return 0, "", err
	}

	// 2. 检查openid是否已存在
	user, err := u.userRepo.GetOrCreateUserByOpenID(openid, req.Nickname, req.Avatar)
	if err != nil {
		return 0, "", err
	}

	// 3. 如果是新用户且提供了位置信息，根据位置分配店铺
	if user.ShopID == 0 && (req.Latitude != 0 || req.Longitude != 0) {
		// 获取最近的店铺
		shopService := NewShopService(u.shopRepo)
		nearestShop, err := shopService.GetNearestShopByLocation(req.Latitude, req.Longitude)
		if err == nil && nearestShop != nil {
			// 更新用户的店铺信息
			updateData := map[string]interface{}{
				"shop_id": nearestShop.ID,
			}
			u.userRepo.UpdateUserByAdmin(user.ID, updateData)
			user.ShopID = nearestShop.ID
		} else {
			// 如果获取店铺失败，默认分配燕郊店
			updateData := map[string]interface{}{
				"shop_id": model.ShopYanjiao,
			}
			u.userRepo.UpdateUserByAdmin(user.ID, updateData)
			user.ShopID = model.ShopYanjiao
		}
	} else if user.ShopID == 0 {
		// 如果没有位置信息，默认分配燕郊店
		updateData := map[string]interface{}{
			"shop_id": model.ShopYanjiao,
		}
		u.userRepo.UpdateUserByAdmin(user.ID, updateData)
		user.ShopID = model.ShopYanjiao
	}

	// 4. 如果用户已存在且已绑定微信，直接返回
	if user.HasWechatBind == model.WechatBindYes {
		token, _ := pkg.GenerateJWTToken(user.ID)
		return user.ID, token, nil
	}

	// 5. 如果用户是新创建的（小程序注册），设置微信绑定状态
	if user.ID > 0 && user.HasWechatBind == model.WechatBindNo {
		err = u.BindWechatToUser(user.ID, openid, req.Nickname)
		if err != nil {
			return 0, "", err
		}
	}

	// 6. 生成自定义 token
	token, _ := pkg.GenerateJWTToken(user.ID)
	return user.ID, token, nil
}

func (u userService) UpdateUserInfo(ctx context.Context, userId int64, req *model.UpdateUserInfoRequest) error {
	return u.userRepo.UpdateUserInfo(userId, req)
}

// CreateUserByAdmin 后台创建用户
func (u userService) CreateUserByAdmin(req *model.AdminUserAddRequest) (*model.User, error) {
	// 检查手机号是否已存在
	_, err := u.userRepo.GetUserByMobilePhone(req.MobilePhone)
	if err == nil {
		return nil, errors.New("手机号已存在")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 如果没有指定店铺，默认分配燕郊店
	shopID := req.ShopID
	if shopID == 0 {
		shopID = model.ShopYanjiao
	}

	user := &model.User{
		AdminDisplayName: req.AdminDisplayName,
		MobilePhone:      req.MobilePhone,
		Source:           model.UserSourceAdmin,
		IsEnable:         model.UserStatusEnabled,
		HasWechatBind:    model.WechatBindNo,
		ShopID:           shopID,
		Remark:           req.Remark,
	}
	err = u.userRepo.CreateUserByAdmin(user)
	return user, err
}

// GetUserByID 根据ID获取用户
func (u userService) GetUserByID(userID int64) (*model.User, error) {
	return u.userRepo.GetUserByID(userID)
}

// UpdateUserByAdmin 后台更新用户信息
func (u userService) UpdateUserByAdmin(req *model.AdminUserEditRequest) error {
	// 检查手机号是否被其他用户使用
	if req.MobilePhone != "" {
		existingUser, err := u.userRepo.GetUserByMobilePhone(req.MobilePhone)
		if err == nil && existingUser.ID != req.ID {
			return errors.New("手机号已被其他用户使用")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) && err != nil {
			return err
		}
	}

	// 构建更新数据
	updateData := map[string]interface{}{}
	if req.AdminDisplayName != "" {
		updateData["admin_display_name"] = req.AdminDisplayName
	}
	if req.MobilePhone != "" {
		updateData["mobile_phone"] = req.MobilePhone
	}
	if req.IsEnable >= 0 {
		updateData["is_enable"] = req.IsEnable
	}
	if req.Remark != "" {
		updateData["remark"] = req.Remark
	}

	return u.userRepo.UpdateUserByAdmin(req.ID, updateData)
}

// GetUserList 获取用户列表
func (u userService) GetUserList(page, pageSize int, keyword string) ([]*model.User, int64, error) {
	return u.userRepo.GetUserList(page, pageSize, keyword)
}

// DeleteUser 删除用户
func (u userService) DeleteUser(userID int64) error {
	return u.userRepo.DeleteUser(userID)
}

// BindWechatToUser 绑定微信到用户
func (u userService) BindWechatToUser(userID int64, openid, wechatDisplayName string) error {
	updateData := map[string]interface{}{
		"openid":              openid,
		"wechat_display_name": wechatDisplayName,
		"has_wechat_bind":     model.WechatBindYes,
		"source":              model.UserSourceMixed,
	}
	return u.userRepo.BindWechatToUser(userID, updateData)
}

// GetUserByMobilePhone 根据手机号获取用户
func (u userService) GetUserByMobilePhone(mobilePhone string) (*model.User, error) {
	return u.userRepo.GetUserByMobilePhone(mobilePhone)
}

// WechatBindMobile 小程序绑定手机号
func (u userService) WechatBindMobile(userID int64, req *model.WechatBindMobileRequest) (*model.User, error) {
	// 检查手机号是否已存在
	existingUser, err := u.userRepo.GetUserByMobilePhone(req.MobilePhone)
	if err == nil {
		// 手机号已存在，绑定到现有用户
		err = u.BindWechatToUser(existingUser.ID, "", "")
		if err != nil {
			return nil, err
		}
		return existingUser, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 手机号不存在，更新当前用户的手机号
	updateData := map[string]interface{}{
		"mobile_phone": req.MobilePhone,
	}
	err = u.userRepo.UpdateUserByAdmin(userID, updateData)
	if err != nil {
		return nil, err
	}

	// 获取更新后的用户信息
	user, err := u.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}
