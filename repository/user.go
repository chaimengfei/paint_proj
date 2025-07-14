package repository

import (
	"cmf/paint_proj/model"
	"errors"
	"gorm.io/gorm"
)

type UserRepository interface {
	GetOrCreateUserByOpenID(openid, nickname, avatar string) (*model.User, error)
	UpdateUserInfo(userId int64, req *model.UpdateUserInfoRequest) error
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
