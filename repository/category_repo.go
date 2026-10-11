package repository

import (
	"OmniGram/domain"

	"gorm.io/gorm"
)

type CategoryRepository struct {
	DB *gorm.DB
}

// GetMainCategories واکشی دسته‌های اصلی (بدون والد)
func (r *CategoryRepository) GetMainCategories(tenantID uint) ([]domain.Category, error) {
	var categories []domain.Category
	err := r.DB.Where("tenant_id = ? AND parent_id IS NULL AND is_active = ?", tenantID, true).Find(&categories).Error
	return categories, err
}

// GetSubCategories واکشی زیردسته‌ها
func (r *CategoryRepository) GetSubCategories(tenantID uint, parentID uint) ([]domain.Category, error) {
	var categories []domain.Category
	err := r.DB.Where("tenant_id = ? AND parent_id = ? AND is_active = ?", tenantID, parentID, true).Find(&categories).Error
	return categories, err
}

func (r *CategoryRepository) GetCategoryByID(tenantID uint, id uint) (*domain.Category, error) {
	var cat domain.Category
	err := r.DB.Where("tenant_id = ? AND id = ?", tenantID, id).First(&cat).Error
	return &cat, err
}

// CreateCategory ساخت دسته‌بندی جدید
func (r *CategoryRepository) CreateCategory(category *domain.Category) error {
	return r.DB.Create(category).Error
}
