package usecase

import (
	"ybg-backend-copy/internal/entity"
	"ybg-backend-copy/internal/repository"
)

type CategoryUsecase interface {
	CreateCategory(c *entity.Category) error
	GetAllCategories() ([]entity.Category, error)
	DeleteCategory(id uint) error
}

type categoryUC struct {
	repo repository.CategoryRepository
}

func NewCategoryUsecase(repo repository.CategoryRepository) CategoryUsecase {
	return &categoryUC{repo: repo}
}

func (u *categoryUC) CreateCategory(c *entity.Category) error      { return u.repo.Create(c) }
func (u *categoryUC) GetAllCategories() ([]entity.Category, error) { return u.repo.GetAll() }
func (u *categoryUC) DeleteCategory(id uint) error                 { return u.repo.Delete(id) }
