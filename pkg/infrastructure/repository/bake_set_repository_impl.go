package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/miu200521358/bone_baker/pkg/domain"
	"github.com/miu200521358/bone_baker/pkg/domain/repository"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
)

type bakeSetRepositoryImpl struct{}

// NewBakeSetRepository BakeSetRepositoryの実装を返す
func NewBakeSetRepository() repository.BakeSetRepository {
	return &bakeSetRepositoryImpl{}
}

func (r *bakeSetRepositoryImpl) Save(bakeSets []*domain.BakeSet, jsonPath string) error {
	if strings.ToLower(filepath.Ext(jsonPath)) != ".json" {
		jsonPath += ".json"
	}

	output, err := json.Marshal(bakeSets)
	if err != nil {
		mlog.E(mi18n.T("物理焼き込みセット保存失敗エラー"), err, "")
		return err
	}

	if err := os.WriteFile(jsonPath, output, 0644); err != nil {
		mlog.E(mi18n.T("物理焼き込みセット保存失敗エラー"), err, "")
		return err
	}

	mlog.I(mi18n.T("物理焼き込みセット保存成功", map[string]any{"Path": jsonPath}))
	return nil
}

func (r *bakeSetRepositoryImpl) Load(jsonPath string) ([]*domain.BakeSet, error) {
	input, err := os.ReadFile(jsonPath)
	if err != nil {
		mlog.E(mi18n.T("物理焼き込みセット読込失敗エラー"), err, "")
		return nil, err
	}

	var bakeSets []*domain.BakeSet
	if err := json.Unmarshal(input, &bakeSets); err != nil {
		mlog.E(mi18n.T("物理焼き込みセット読込失敗エラー"), err, "")
		return nil, err
	}

	mlog.I(mi18n.T("物理焼き込みセット読込成功", map[string]any{"Path": jsonPath}))
	return bakeSets, nil
}

// GetByID 指定されたIDのBakeSetを取得（簡易実装）
func (r *bakeSetRepositoryImpl) GetByID(id int) (*domain.BakeSet, error) {
	// 実際の実装では、永続化ストレージから取得する
	// ここでは簡易実装として、新しいBakeSetを返す
	return domain.NewPhysicsSet(id), nil
}

// SaveSingle 単一のBakeSetを保存（簡易実装）
func (r *bakeSetRepositoryImpl) SaveSingle(bakeSet *domain.BakeSet) error {
	// 実際の実装では、永続化ストレージに保存する
	// ここでは簡易実装として、何もしない
	return nil
}
