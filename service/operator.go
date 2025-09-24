package service

import (
	"cmf/paint_proj/model"
	"cmf/paint_proj/pkg"
	"cmf/paint_proj/repository"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type OperatorService interface {
	// 管理员登录
	AdminLogin(req *model.AdminLoginRequest) (*model.AdminLoginResponse, error)

	// 根据ID获取管理员
	GetOperatorByID(operatorID int64) (*model.Operator, error)

	// 获取管理员列表
	GetOperatorList(page, pageSize int, keyword string) ([]*model.Operator, int64, error)
}

type operatorService struct {
	operatorRepo repository.OperatorRepository
	shopRepo     repository.ShopRepository
}

func NewOperatorService(operatorRepo repository.OperatorRepository, shopRepo repository.ShopRepository) OperatorService {
	return &operatorService{
		operatorRepo: operatorRepo,
		shopRepo:     shopRepo,
	}
}

func (s *operatorService) AdminLogin(req *model.AdminLoginRequest) (*model.AdminLoginResponse, error) {
	// 1. 根据账号获取管理员
	operator, err := s.operatorRepo.GetOperatorByUsername(req.OperatorName)
	if err != nil {
		return nil, errors.New("账号或密码错误")
	}

	// 2. 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(operator.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("账号或密码错误")
	}

	// 3. 生成 JWT Token
	isRoot := operator.Name == "root" // root 账号为超级管理员
	token, err := pkg.GenerateAdminJWTToken(operator.ID, operator.Name, operator.ShopID, isRoot)
	if err != nil {
		return nil, errors.New("生成 token 失败")
	}

	// 4. 根据管理员类型返回不同的店铺信息
	response := &model.AdminLoginResponse{
		Token:     token,
		Operator:  operator,
		ExpiresIn: 7200, // 2小时
	}

	if isRoot {
		// 超级管理员返回所有店铺信息
		shops, err := s.shopRepo.GetAllActiveShops()
		if err != nil {
			return nil, errors.New("获取店铺列表失败")
		}

		// 转换为简化的店铺信息
		shopList := make([]model.ShopSimple, 0, len(shops))
		for _, shop := range shops {
			shopSimple := model.ShopSimple{
				ID:          shop.ID,
				Name:        shop.Name,
				Code:        shop.Code,
				Address:     shop.Address,
				Phone:       shop.Phone,
				ManagerName: shop.ManagerName,
				IsActive:    shop.IsActive,
			}
			shopList = append(shopList, shopSimple)
		}
		response.ShopList = shopList
	} else {
		// 普通管理员返回单个店铺信息
		shop, err := s.shopRepo.GetShopByID(operator.ShopID)
		if err != nil {
			return nil, errors.New("获取店铺信息失败")
		}

		shopSimple := &model.ShopSimple{
			ID:          shop.ID,
			Name:        shop.Name,
			Code:        shop.Code,
			Address:     shop.Address,
			Phone:       shop.Phone,
			ManagerName: shop.ManagerName,
			IsActive:    shop.IsActive,
		}
		response.ShopInfo = shopSimple
	}

	return response, nil
}

func (s *operatorService) GetOperatorByID(operatorID int64) (*model.Operator, error) {
	return s.operatorRepo.GetOperatorByID(operatorID)
}

func (s *operatorService) GetOperatorList(page, pageSize int, keyword string) ([]*model.Operator, int64, error) {
	return s.operatorRepo.GetOperatorList(page, pageSize, keyword)
}
