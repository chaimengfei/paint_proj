package service

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/repository"
	"errors"
)

type AddressService interface {
	GetAddressList(userID int64) ([]*model.AddressInfo, error)
	CreateAddress(userID int64, req *model.CreateAddressReq) error
	SetDefaultAddress(userID, addressId int64) error
	UpdateAddress(userID, addressId int64, req *model.UpdateAddressReq) error
	DeleteAddress(userID, addressId int64) error

	GetAdminAddressList(userId int64, userName string) ([]*model.AdminAddressInfo, error)
	CreateAdminAddress(userId int64, req *model.CreateAddressReq) error
	UpdateAdminAddress(addressId int64, req *model.UpdateAddressReq) error
}

type addressService struct {
	addressRepo repository.AddressRepository
}

func NewAddressService(ar repository.AddressRepository) AddressService {
	return &addressService{
		addressRepo: ar,
	}
}

func (a addressService) GetAddressList(userID int64) ([]*model.AddressInfo, error) {
	dbList, err := a.addressRepo.GetByUserId(userID)
	if err != nil {
		return nil, err
	}
	res := make([]*model.AddressInfo, 0, len(dbList))
	for _, dbData := range dbList {
		isDefault := dbData.IsDefault == 1
		res = append(res, &model.AddressInfo{
			AddressID:      dbData.ID,
			RecipientName:  dbData.RecipientName,
			RecipientPhone: dbData.RecipientPhone,
			Province:       dbData.Province,
			City:           dbData.City,
			District:       dbData.District,
			Detail:         dbData.Detail,
			IsDefault:      &isDefault,
		})
	}
	return res, nil
}

func (a addressService) CreateAddress(userID int64, req *model.CreateAddressReq) error {
	dbData := model.Address{
		UserId:         userID,
		RecipientName:  req.Data.RecipientName,
		RecipientPhone: req.Data.RecipientPhone,
		Province:       req.Data.Province,
		City:           req.Data.City,
		District:       req.Data.District,
		Detail:         req.Data.Detail,
	}
	// 如果设置为默认，则取消用户其他地址的默认状态
	if req.Data.IsDefault != nil {
		if *req.Data.IsDefault {
			dbData.IsDefault = 1 // 设置默认
		} else {
			dbData.IsDefault = 0 // 取消默认
		}
	} else {
		dbData.IsDefault = 0 // 忽略默认地址设置
	}
	return a.addressRepo.Create(&dbData)
}

func (a addressService) SetDefaultAddress(userID, addressId int64) error {
	return a.addressRepo.SetDefault(userID, addressId)
}
func (a addressService) UpdateAddress(userID, addressId int64, req *model.UpdateAddressReq) error {
	dbData := map[string]interface{}{}
	if req.Data.RecipientName != "" {
		dbData["recipient_name"] = req.Data.RecipientName
	}
	if req.Data.RecipientPhone != "" {
		dbData["recipient_phone"] = req.Data.RecipientPhone
	}
	if req.Data.Province != "" {
		dbData["province"] = req.Data.Province
	}
	if req.Data.City != "" {
		dbData["city"] = req.Data.City
	}
	if req.Data.District != "" {
		dbData["district"] = req.Data.District
	}
	if req.Data.Detail != "" {
		dbData["detail"] = req.Data.Detail
	}
	// 如果设置为默认，则取消用户其他地址的默认状态
	if req.Data.IsDefault != nil {
		if *req.Data.IsDefault {
			dbData["is_default"] = 1 // 设置默认
		} else {
			dbData["is_default"] = 0 // 取消默认
		}
	} else {
		dbData["is_default"] = 0 // 忽略默认地址设置
	}
	return a.addressRepo.Update(addressId, dbData)
}

func (a addressService) DeleteAddress(userID, addressId int64) error {
	return a.addressRepo.Delete(addressId)
}

// GetAdminAddressList 获取admin地址列表
func (a addressService) GetAdminAddressList(userId int64, userName string) ([]*model.AdminAddressInfo, error) {
	dbList, err := a.addressRepo.GetAddressListByUser(userId, userName)
	if err != nil {
		return nil, err
	}

	res := make([]*model.AdminAddressInfo, 0, len(dbList))
	for _, dbData := range dbList {
		res = append(res, &model.AdminAddressInfo{
			AddressID:      dbData.ID,
			UserID:         dbData.UserId,
			UserName:       dbData.UserName,
			RecipientName:  dbData.RecipientName,
			RecipientPhone: dbData.RecipientPhone,
			Province:       dbData.Province,
			City:           dbData.City,
			District:       dbData.District,
			Detail:         dbData.Detail,
			IsDefault:      dbData.IsDefault == 1,
			CreatedAt:      "", // Address模型中没有CreatedAt字段
		})
	}
	return res, nil
}

func (a addressService) CreateAdminAddress(userId int64, req *model.CreateAddressReq) error {
	// 检查请求数据是否包含用户ID
	if userId == 0 {
		return errors.New("用户ID不能为空")
	}

	dbData := model.Address{
		UserId:         userId,
		RecipientName:  req.Data.RecipientName,
		RecipientPhone: req.Data.RecipientPhone,
		Province:       req.Data.Province,
		City:           req.Data.City,
		District:       req.Data.District,
		Detail:         req.Data.Detail,
	}

	// 如果设置为默认，则取消用户其他地址的默认状态
	if req.Data.IsDefault != nil {
		if *req.Data.IsDefault {
			dbData.IsDefault = 1 // 设置默认
		} else {
			dbData.IsDefault = 0 // 取消默认
		}
	} else {
		dbData.IsDefault = 0 // 默认不设置为默认地址
	}

	return a.addressRepo.Create(&dbData)
}

func (a addressService) UpdateAdminAddress(addressId int64, req *model.UpdateAddressReq) error {
	// 先检查地址是否存在
	existingAddress, err := a.addressRepo.GetById(addressId)
	if err != nil {
		return err
	}
	if existingAddress == nil {
		return errors.New("地址不存在")
	}

	dbData := map[string]interface{}{}
	if req.Data.RecipientName != "" {
		dbData["recipient_name"] = req.Data.RecipientName
	}
	if req.Data.RecipientPhone != "" {
		dbData["recipient_phone"] = req.Data.RecipientPhone
	}
	if req.Data.Province != "" {
		dbData["province"] = req.Data.Province
	}
	if req.Data.City != "" {
		dbData["city"] = req.Data.City
	}
	if req.Data.District != "" {
		dbData["district"] = req.Data.District
	}
	if req.Data.Detail != "" {
		dbData["detail"] = req.Data.Detail
	}

	// 如果设置为默认，则取消用户其他地址的默认状态
	if req.Data.IsDefault != nil {
		if *req.Data.IsDefault {
			dbData["is_default"] = 1 // 设置默认
		} else {
			dbData["is_default"] = 0 // 取消默认
		}
	}

	return a.addressRepo.Update(addressId, dbData)
}
