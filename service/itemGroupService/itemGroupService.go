package itemGroupService

import (
	"Bea-Cukai/model"
	"Bea-Cukai/repo/itemGroupRepository"
	"context"
)

// ItemGroupService sits on top of the itemGroupRepository and exposes use-case oriented APIs.

type ItemGroupService struct {
	itemGroupRepo *itemGroupRepository.ItemGroupRepository
}

func NewItemGroupService(itemGroupRepo *itemGroupRepository.ItemGroupRepository) *ItemGroupService {
	return &ItemGroupService{itemGroupRepo: itemGroupRepo}
}

// ==========================
// Business Operations
// ==========================

// GetAll retrieves all item groups
func (s *ItemGroupService) GetAll() ([]model.ItemGroup, error) {
	ctx := context.Background()
	return s.itemGroupRepo.GetAll(ctx)
}
