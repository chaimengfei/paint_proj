package repository

import (
	"cmf/paint_proj/model"
	"errors"
	"gorm.io/gorm"
)

type AddressRepository interface {
	GetById(id int64) (*model.Address, error)
	GetByUserId(userId int64) ([]model.Address, error)
	GetByUserAppointId(userId, id int64) (*model.Address, error)
	GetDefaultOrFirstAddressID(userId int64) (*model.Address, error)

	Create(data *model.Address) error
	Update(id int64, data map[string]interface{}) error
	SetDefault(userId, id int64) error
	Delete(id int64) error
}

type addressRepository struct {
	db *gorm.DB
}

func NewAddressRepository(db *gorm.DB) AddressRepository {
	return &addressRepository{db: db}
}

func (ar *addressRepository) GetById(id int64) (*model.Address, error) {
	var address model.Address
	err := ar.db.Model(&model.Address{}).Where("id = ?", id).First(&address).Error
	return &address, err
}

func (ar *addressRepository) GetByUserId(userId int64) ([]model.Address, error) {
	var result []model.Address
	err := ar.db.Model(&model.Address{}).Where("user_id = ? AND is_delete = 0", userId).Order("is_default desc").Find(&result).Error
	return result, err
}

func (ar *addressRepository) GetByUserAppointId(userId, id int64) (*model.Address, error) {
	var address model.Address
	err := ar.db.Model(&model.Address{}).Where("user_id = ? and id = ?", userId, id).First(&address).Error
	return &address, err
}

func (ar *addressRepository) GetDefaultOrFirstAddressID(userId int64) (*model.Address, error) {
	var address model.Address
	err := ar.db.Model(&model.Address{}).Where("user_id = ? AND is_default = 1 AND is_delete = 0", userId).First(&address).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = ar.db.Model(&model.Address{}).Where("user_id = ? AND is_delete = 0", userId).Order("id asc").First(&address).Error
	}
	return &address, err
}

func (ar *addressRepository) Create(data *model.Address) error {
	return ar.db.Model(&model.Address{}).Create(data).Error
}

func (ar *addressRepository) Update(id int64, data map[string]interface{}) error {
	return ar.db.Model(&model.Address{}).Where("id = ?", id).Updates(&data).Error
}
func (ar *addressRepository) SetDefault(userId, id int64) error {
	err := ar.db.Transaction(func(tx *gorm.DB) error {
		// 将现有的1设置为0
		err := tx.Model(&model.Address{}).Where("user_id = ? and is_delete = 1", userId).Updates(map[string]interface{}{"is_default": 0}).Error
		if err != nil {
			return err
		}
		// 将当前的0设置为1
		err = tx.Model(&model.Address{}).Where("user_id = ? and id = ?", id).Updates(map[string]interface{}{"is_default": 1}).Error
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (ar *addressRepository) Delete(id int64) error {
	return ar.db.Model(&model.Address{}).Where("id = ?", id).Updates(&model.Address{IsDelete: 1}).Error
}
