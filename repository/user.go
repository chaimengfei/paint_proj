package repository

import (
	"cmf/paint_proj/model"
	"errors"

	"gorm.io/gorm"
)

type UserRepository interface {
	GetOrCreateUserByOpenID(openid, nickname, avatar string) (*model.User, error)
	UpdateUserInfo(userId int64, req *model.UpdateUserInfoRequest) error

	// 后台用户管理
	CreateUserByAdmin(user *model.User) error
	UpdateUserByAdmin(userID int64, updateData map[string]interface{}) error

	GetUserByID(userID int64) (*model.User, error)
	GetUserByMobilePhone(mobilePhone string) (*model.User, error)
	GetUserList(page, pageSize int, keyword string) ([]*model.User, int64, error)
	GetUserListByShop(page, pageSize int, keyword string, shopID int64) ([]*model.User, int64, error)
	DeleteUser(userID int64) error

	// 小程序用户绑定
	BindWechatToUser(userID int64, updateData map[string]interface{}) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}
func (u userRepository) GetOrCreateUserByOpenID(openid, nickname, avatar string) (*model.User, error) {
	var user model.User
	err := u.db.Where("openid = ?", openid).First(&user).Error
	if err == nil {
		return &user, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		user = model.User{
			Openid:   openid,
			Nickname: nickname,
			Avatar:   avatar,
		}
		if err = u.db.Create(&user).Error; err != nil {
			return nil, err
		}
		return &user, nil
	}
	return nil, err
}

func (u userRepository) UpdateUserInfo(userId int64, req *model.UpdateUserInfoRequest) error {
	updateVal := map[string]interface{}{}
	if req.Nickname != "" {
		updateVal["nickname"] = req.Nickname
	}
	if req.Mobile != "" {
		updateVal["mobile_phone"] = req.Mobile
	}
	err := u.db.Model(&model.User{}).Where("id = ?", userId).Updates(updateVal).Error
	return err
}

// CreateUserByAdmin 后台创建用户
func (u userRepository) CreateUserByAdmin(user *model.User) error {
	return u.db.Model(&model.User{}).Create(user).Error
}

// GetUserByID 根据ID获取用户
func (u userRepository) GetUserByID(userID int64) (*model.User, error) {
	var user model.User
	err := u.db.Where("id = ?", userID).First(&user).Error
	return &user, err
}

// GetUserByMobilePhone 根据手机号获取用户
func (u userRepository) GetUserByMobilePhone(mobilePhone string) (*model.User, error) {
	var user model.User
	err := u.db.Model(&model.User{}).Where("mobile_phone = ?", mobilePhone).First(&user).Error
	return &user, err
}

// UpdateUserByAdmin 后台更新用户信息
func (u userRepository) UpdateUserByAdmin(userID int64, updateData map[string]interface{}) error {
	return u.db.Model(&model.User{}).Where("id = ?", userID).Updates(updateData).Error
}

// GetUserList 获取用户列表
func (u userRepository) GetUserList(page, pageSize int, keyword string) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	query := u.db.Model(&model.User{})

	// 搜索条件
	if keyword != "" {
		query = query.Where("admin_display_name LIKE ? OR mobile_phone LIKE ? OR wechat_display_name LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&users).Error
	return users, total, err
}

// GetUserListByShop 根据店铺获取用户列表
func (u userRepository) GetUserListByShop(page, pageSize int, keyword string, shopID int64) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	query := u.db.Model(&model.User{}).Where("shop_id = ?", shopID)

	// 搜索条件
	if keyword != "" {
		query = query.Where("admin_display_name LIKE ? OR mobile_phone LIKE ? OR wechat_display_name LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&users).Error
	return users, total, err
}

// DeleteUser 删除用户
func (u userRepository) DeleteUser(userID int64) error {
	return u.db.Model(&model.User{}).Delete(&model.User{}, userID).Error
}

// BindWechatToUser 绑定微信到用户
func (u userRepository) BindWechatToUser(userID int64, updateData map[string]interface{}) error {
	return u.db.Model(&model.User{}).Where("id = ?", userID).Updates(updateData).Error
}
