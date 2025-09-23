package repository

import (
	"cmf/paint_proj/model"

	"gorm.io/gorm"
)

type OperatorRepository interface {
	// 根据账号获取管理员
	GetOperatorByUsername(username string) (*model.Operator, error)

	// 根据ID获取管理员
	GetOperatorByID(operatorID int64) (*model.Operator, error)

	// 获取管理员列表
	GetOperatorList(page, pageSize int, keyword string) ([]*model.Operator, int64, error)
}

type operatorRepository struct {
	db *gorm.DB
}

func NewOperatorRepository(db *gorm.DB) OperatorRepository {
	return &operatorRepository{
		db: db,
	}
}

func (r *operatorRepository) GetOperatorByUsername(username string) (*model.Operator, error) {
	var operator model.Operator
	err := r.db.Where("name = ? AND is_active = 1", username).First(&operator).Error
	if err != nil {
		return nil, err
	}
	return &operator, nil
}

func (r *operatorRepository) GetOperatorByID(operatorID int64) (*model.Operator, error) {
	var operator model.Operator
	err := r.db.Where("id = ? AND is_active = 1", operatorID).First(&operator).Error
	if err != nil {
		return nil, err
	}
	return &operator, nil
}

func (r *operatorRepository) GetOperatorList(page, pageSize int, keyword string) ([]*model.Operator, int64, error) {
	var operators []*model.Operator
	var total int64

	query := r.db.Model(&model.Operator{}).Where("is_active = 1")

	if keyword != "" {
		query = query.Where("name LIKE ? OR real_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&operators).Error; err != nil {
		return nil, 0, err
	}

	return operators, total, nil
}
