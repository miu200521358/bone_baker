package usecase

import (
	"github.com/miu200521358/bone_baker/pkg/domain/entity"
	pRepository "github.com/miu200521358/bone_baker/pkg/infrastructure/repository"
)

type SaveUsecase struct {
	fileRepo *pRepository.FileRepository
}

func NewSaveUsecase(fileRepo *pRepository.FileRepository) *SaveUsecase {
	return &SaveUsecase{
		fileRepo: fileRepo,
	}
}

func (uc *SaveUsecase) SaveFile(bakeSets []*entity.BakeSet, physicsRecords []*entity.PhysicsRecord, path string) error {
	return uc.fileRepo.Save(bakeSets, physicsRecords, path)
}
