package service

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/repository"
)

type AddressService interface {
	GetAddressList(userID int64) ([]*model.AddressInfo, error)
	CreateAddress(userID int64, req *model.CreateAddressReq) error
	SetDefaultAddress(userID, addressId int64) error
	UpdateAddress(userID, addressId int64, req *model.UpdateAddressReq) error
	DeleteAddress(userID, addressId int64) error
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
