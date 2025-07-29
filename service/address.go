package service

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/repository"
)

type AddressService interface {
	GetAddressList(userID int64) ([]*model.Address, error)
	CreateAddress(data *model.Address) error
	SetDefaultAddress(userID, addressId int64) error
	UpdateAddress(userID, addressId int64, data map[string]interface{}) error
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

func (a addressService) GetAddressList(userID int64) ([]*model.Address, error) {
	return a.addressRepo.GetByUserId(userID)
}

func (a addressService) CreateAddress(data *model.Address) error {
	return a.addressRepo.Create(data)
}

func (a addressService) SetDefaultAddress(userID, addressId int64) error {
	return a.addressRepo.SetDefault(userID, addressId)
}
func (a addressService) UpdateAddress(userID, addressId int64, data map[string]interface{}) error {
	return a.addressRepo.Update(addressId, data)
}

func (a addressService) DeleteAddress(userID, addressId int64) error {
	return a.addressRepo.Delete(addressId)
}
