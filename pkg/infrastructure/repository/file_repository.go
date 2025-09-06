package infrastructure

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/miu200521358/bone_baker/pkg/domain/entity"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
)

type FileRepository struct{}

// NewFileRepository コンストラクタ
func NewFileRepository() *FileRepository {
	return &FileRepository{}
}

// Save BakeSetのリストをJSONファイルに保存
func (r *FileRepository) Save(bakeSets []*entity.BakeSet, filePath string) error {
	// ファイル拡張子の確認
	if strings.ToLower(filepath.Ext(filePath)) != ".json" {
		filePath += ".json"
	}

	// JSONにシリアライズ
	output, err := json.Marshal(bakeSets)
	if err != nil {
		mlog.E(mi18n.T("物理焼き込みセット保存失敗エラー"), err, "")
		return err
	}

	// ファイルに書き込み
	if err := os.WriteFile(filePath, output, 0644); err != nil {
		mlog.E(mi18n.T("物理焼き込みセット保存失敗エラー"), err, "")
		return err
	}

	mlog.I(mi18n.T("物理焼き込みセット保存成功", map[string]any{"Path": filePath}))
	return nil
}

// Load JSONファイルからBakeSetのリストを読み込み
func (r *FileRepository) Load(filePath string) ([]*entity.BakeSet, error) {
	// ファイル読み込み
	input, err := os.ReadFile(filePath)
	if err != nil {
		mlog.E(mi18n.T("物理焼き込みセット読込失敗エラー"), err, "")
		return nil, err
	}

	// JSONから逆シリアライズ
	var bakeSets []*entity.BakeSet
	if err := json.Unmarshal(input, &bakeSets); err != nil {
		mlog.E(mi18n.T("物理焼き込みセット読込失敗エラー"), err, "")
		return nil, err
	}

	mlog.I(mi18n.T("物理焼き込みセット読込成功", map[string]any{"Path": filePath}))
	return bakeSets, nil
}
