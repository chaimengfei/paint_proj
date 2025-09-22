package service

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/repository"
	"math"
)

type ShopService interface {
	// 根据地理位置获取最近的店铺
	GetNearestShopByLocation(latitude, longitude float64) (*model.Shop, error)

	// 获取所有启用的店铺
	GetAllActiveShops() ([]*model.Shop, error)

	// 根据ID获取店铺
	GetShopByID(shopID int64) (*model.Shop, error)

	// 计算两点间距离（公里）
	CalculateDistance(lat1, lon1, lat2, lon2 float64) float64
}

type shopService struct {
	shopRepo repository.ShopRepository
}

func NewShopService(sr repository.ShopRepository) ShopService {
	return &shopService{
		shopRepo: sr,
	}
}

// GetNearestShopByLocation 根据地理位置获取最近的店铺
func (s *shopService) GetNearestShopByLocation(latitude, longitude float64) (*model.Shop, error) {
	// 获取所有启用的店铺
	shops, err := s.shopRepo.GetAllActiveShops()
	if err != nil {
		return nil, err
	}

	if len(shops) == 0 {
		// 如果没有店铺，返回默认的燕郊店
		return s.shopRepo.GetShopByID(model.ShopYanjiao)
	}

	var nearestShop *model.Shop
	minDistance := math.MaxFloat64

	for _, shop := range shops {
		distance := s.CalculateDistance(latitude, longitude, shop.Latitude, shop.Longitude)
		if distance < minDistance {
			minDistance = distance
			nearestShop = shop
		}
	}

	// 如果最近距离超过50公里，默认返回燕郊店
	if minDistance > 50 {
		return s.shopRepo.GetShopByID(model.ShopYanjiao)
	}

	return nearestShop, nil
}

// CalculateDistance 计算两点间距离（公里）
func (s *shopService) CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // 地球半径（公里）

	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

func (s *shopService) GetAllActiveShops() ([]*model.Shop, error) {
	return s.shopRepo.GetAllActiveShops()
}

func (s *shopService) GetShopByID(shopID int64) (*model.Shop, error) {
	return s.shopRepo.GetShopByID(shopID)
}
