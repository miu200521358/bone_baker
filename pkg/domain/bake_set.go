package domain

import (
	"errors"
	"fmt"

	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/domain/mmath"
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
	"github.com/miu200521358/mlib_go/pkg/infrastructure/mfile"
)

type BakeSet struct {
	Index       int  // インデックス
	IsTerminate bool // 処理停止フラグ

	// Value Objectsを使用したファイルパス
	originalMotionPath *FilePath `json:"-"` // 元モーションパス（Value Object）
	originalModelPath  *FilePath `json:"-"` // 元モデルパス（Value Object）
	outputMotionPath   *FilePath `json:"-"` // 出力モーションパス（Value Object）
	outputModelPath    *FilePath `json:"-"` // 出力モデルパス（Value Object）

	// JSONシリアライズ用の文字列フィールド（後方互換性）
	OriginalMotionPathStr string `json:"original_motion_path"` // 元モーションパス
	OriginalModelPathStr  string `json:"original_model_path"`  // 元モデルパス

	OriginalMotionName string `json:"-"` // 元モーション名
	OriginalModelName  string `json:"-"` // 元モーション名
	OutputModelName    string `json:"-"` // 物理焼き込み先モデル名

	OriginalMotion *vmd.VmdMotion `json:"-"` // 元モデル
	OriginalModel  *pmx.PmxModel  `json:"-"` // 元モデル
	BakedModel     *pmx.PmxModel  `json:"-"` // 物理焼き込み先モデル
	OutputMotion   *vmd.VmdMotion `json:"-"` // 出力結果モーション

	PhysicsTableModel *PhysicsTableModel `json:"physics_table"` // 物理ボーンツリー
	OutputTableModel  *OutputTableModel  `json:"output_table"`  // 出力定義テーブル
}

func NewPhysicsSet(index int) *BakeSet {
	return &BakeSet{
		Index:              index,
		PhysicsTableModel:  NewPhysicsTableModel(),
		OutputTableModel:   NewOutputTableModel(),
		originalMotionPath: NewFilePath(""),
		originalModelPath:  NewFilePath(""),
		outputMotionPath:   NewFilePath(""),
		outputModelPath:    NewFilePath(""),
	}
}

// Getter methods for Value Objects
func (ss *BakeSet) OriginalMotionPath() string {
	if ss.originalMotionPath == nil {
		return ""
	}
	return ss.originalMotionPath.Value()
}

func (ss *BakeSet) OriginalModelPath() string {
	if ss.originalModelPath == nil {
		return ""
	}
	return ss.originalModelPath.Value()
}

func (ss *BakeSet) OutputMotionPath() string {
	if ss.outputMotionPath == nil {
		return ""
	}
	return ss.outputMotionPath.Value()
}

func (ss *BakeSet) OutputModelPath() string {
	if ss.outputModelPath == nil {
		return ""
	}
	return ss.outputModelPath.Value()
}

// Setter methods for Value Objects
func (ss *BakeSet) SetOriginalMotionPath(path string) {
	ss.originalMotionPath = NewFilePath(path)
	ss.OriginalMotionPathStr = path
}

func (ss *BakeSet) SetOriginalModelPath(path string) {
	ss.originalModelPath = NewFilePath(path)
	ss.OriginalModelPathStr = path
}

func (ss *BakeSet) SetOutputMotionPath(path string) {
	ss.outputMotionPath = NewFilePath(path)
}

func (ss *BakeSet) SetOutputModelPath(path string) {
	ss.outputModelPath = NewFilePath(path)
}

func (ss *BakeSet) MaxFrame() float32 {
	if ss.OriginalMotion == nil {
		return 0
	}

	return ss.OriginalMotion.MaxFrame()
}
func (ss *BakeSet) CreateOutputModelPath() string {
	if ss.OriginalModel == nil {
		return ""
	}

	// 物理焼き込み先モデルが指定されている場合、ファイル名を含める
	return mfile.CreateOutputPath(ss.OriginalModel.Path(), "BB")
}

func (ss *BakeSet) CreateOutputMotionPath() string {
	if ss.OriginalMotion == nil || ss.BakedModel == nil {
		return ""
	}

	// 物理焼き込み先モデルが指定されている場合、ファイル名を含める
	_, fileName, _ := mfile.SplitPath(ss.BakedModel.Path())

	return mfile.CreateOutputPath(
		ss.OriginalMotion.Path(), fmt.Sprintf("BB_%s", fileName))
}

func (ss *BakeSet) setMotion(originalMotion, outputMotion *vmd.VmdMotion) {
	if originalMotion == nil || outputMotion == nil {
		ss.SetOriginalMotionPath("")
		ss.OriginalMotionName = ""
		ss.OriginalMotion = nil

		ss.SetOutputMotionPath("")
		ss.OutputMotion = vmd.NewVmdMotion("")

		return
	}

	ss.OriginalMotionName = originalMotion.Name()
	ss.OriginalMotion = originalMotion
	ss.OutputMotion = outputMotion
}

func (ss *BakeSet) setModels(originalModel, physicsBakedModel *pmx.PmxModel) {
	if originalModel == nil {
		ss.SetOriginalModelPath("")
		ss.OriginalModelName = ""
		ss.OriginalModel = nil
		ss.BakedModel = nil
		return
	}

	ss.SetOriginalModelPath(originalModel.Path())
	ss.OriginalModelName = originalModel.Name()
	ss.OriginalModel = originalModel
	ss.BakedModel = physicsBakedModel
}

// SetModels ドメインロジックでモデルを設定（公開メソッド）
func (ss *BakeSet) SetModels(originalModel, bakedModel *pmx.PmxModel) error {
	if originalModel == nil {
		ss.setModels(nil, nil)
		return nil
	}

	// ドメインサービスを使用
	physicsBoneService := NewPhysicsBoneService()

	// ドメインルールの適用
	physicsBoneService.ProcessPhysicsBones(originalModel)
	physicsBoneService.ProcessPhysicsBones(bakedModel)

	if bakedModel != nil {
		physicsBoneService.FixPhysicsRigidBodies(bakedModel)
	}

	ss.setModels(originalModel, bakedModel)
	ss.SetOutputModelPath(ss.CreateOutputModelPath())

	return nil
}

// ClearModels モデルをクリア（公開メソッド）
func (ss *BakeSet) ClearModels() {
	ss.setModels(nil, nil)
}

// SetMotions ドメインロジックでモーションを設定（公開メソッド）
func (ss *BakeSet) SetMotions(originalMotion, outputMotion *vmd.VmdMotion) error {
	ss.setMotion(originalMotion, outputMotion)
	ss.SetOutputMotionPath(ss.CreateOutputMotionPath())
	return nil
}

// ClearMotions モーションをクリア（公開メソッド）
func (ss *BakeSet) ClearMotions() {
	ss.setMotion(nil, nil)
}

func (ss *BakeSet) Delete() {
	ss.SetOriginalMotionPath("")
	ss.SetOriginalModelPath("")
	ss.SetOutputMotionPath("")
	ss.SetOutputModelPath("")

	ss.OriginalMotionName = ""
	ss.OriginalModelName = ""
	ss.OutputModelName = ""

	ss.OriginalMotion = nil
	ss.OriginalModel = nil
	ss.OutputMotion = nil
}

// 物理ボーンだけ残す
func (ss *BakeSet) GetOutputMotionOnlyChecked(records []*OutputBoneRecord) ([]*vmd.VmdMotion, error) {
	motions := make([]*vmd.VmdMotion, 0)

	if ss.OriginalModel == nil || ss.OutputMotion == nil || len(records) == 0 {
		return motions, errors.New(mi18n.T("物理焼き込みセットの元モデルまたは出力モーションが設定されていません"))
	}

	var bakedMotion *vmd.VmdMotion

	// まずは既存モーションに焼き込みボーンを追加挿入する
	var err error
	bakedMotion, err = ss.OriginalMotion.Copy()
	bakedMotion.SetPath(ss.OutputMotionPath())
	if err != nil {
		return motions, fmt.Errorf(mi18n.T("元モーションのコピーに失敗しました: %w"), err)
	}

	keyCounts := make([]int, int(ss.OriginalMotion.MaxFrame()+1+1))
	for _, record := range records {
		if record == nil || record.OutputBoneTreeModel == nil {
			continue
		}

		for f := record.StartFrame; f <= record.EndFrame; f++ {
			for _, boneName := range record.TargetBoneNames {
				bf := ss.OutputMotion.BoneFrames.Get(boneName).Get(f)

				if bf == nil {
					continue
				}

				if bone, err := ss.OriginalModel.Bones.GetByName(boneName); err == nil {
					if bone.HasPhysics() {
						bf.DisablePhysics = true
					}
					bakedMotion.AppendBoneFrame(boneName, bf)
					keyCounts[int(f)]++

					// 次のキーフレ物理有効で登録しておく
					if bone.HasPhysics() {
						bf := ss.OutputMotion.BoneFrames.Get(boneName).Get(f + 1)
						if bf == nil {
							continue
						}
						bf.DisablePhysics = false
						bakedMotion.AppendBoneFrame(boneName, bf)
						keyCounts[int(f)]++
					}
				}
			}
		}
	}

	if mmath.Sum(keyCounts) == 0 {
		return motions, errors.New(mi18n.T("焼き込み対象キーフレームなし"))
	}

	dirPath, fileName, ext := mfile.SplitPath(ss.OutputMotionPath())
	motion := vmd.NewVmdMotion("")
	motion.SetPath(fmt.Sprintf("%s%s_%04d%s", dirPath, fileName, 0, ext))
	motion.MorphFrames, _ = ss.OriginalMotion.MorphFrames.Copy()

	frameCount := 0
	logFrameCount := 0

	for f := 0; f <= len(keyCounts); f++ {
		if f < len(keyCounts)-1 && frameCount+keyCounts[int(f+1)] > vmd.MAX_BONE_FRAMES {
			// キーフレーム数が上限を超える場合は切り替える
			motions = append(motions, motion)

			dirPath, fileName, ext := mfile.SplitPath(ss.OutputMotionPath())
			motion = vmd.NewVmdMotion(fmt.Sprintf("%s%s_%04d%s", dirPath, fileName, f, ext))
			motion.MorphFrames, _ = ss.OriginalMotion.MorphFrames.Copy()

			mlog.I(fmt.Sprintf(mi18n.T("キーフレーム数が上限を超えるため、モーションを切り替えます[%04dF]: %d -> %d"),
				f, frameCount, vmd.MAX_BONE_FRAMES))

			frameCount = 0
			logFrameCount = 0
		}

		if frameCount/100000 > logFrameCount/100000 {
			mlog.I(fmt.Sprintf(mi18n.T("- 物理焼き込み中... [%04dF] %dキーフレーム"), f, frameCount))
			logFrameCount = frameCount
		}

		ss.OriginalModel.Bones.ForEach(func(boneIndex int, bone *pmx.Bone) bool {
			if bakedMotion.BoneFrames.Get(bone.Name()).Contains(float32(f)) {
				// キーフレームがある場合は登録
				motion.AppendBoneFrame(bone.Name(), bakedMotion.BoneFrames.Get(bone.Name()).Get(float32(f)))
			}

			return true
		})

		if f < len(keyCounts)-1 {
			frameCount += keyCounts[int(f)]
		}
	}

	motions = append(motions, motion)
	return motions, nil
}

// // SetOutputChildrenChecked は指定されたアイテムの子どもを再帰的にチェック状態を設定する
// func (ss *BakeSet) SetOutputChildrenChecked(treeView *walk.TreeView, item walk.TreeItem, checked bool) {
// 	if item == nil || ss.IsOutputUpdatingChildren ||
// 		ss.IsOutputUpdatingPhysics || ss.IsOutputUpdatingIk {
// 		return
// 	}

// 	// 無限ループを防ぐためのフラグ
// 	ss.IsOutputUpdatingChildren = true
// 	defer func() {
// 		ss.IsOutputUpdatingChildren = false
// 	}()

// 	item.(*OutputItem).SetChecked(checked)

// 	// 子どもの数を取得
// 	childCount := item.ChildCount()
// 	for i := range childCount {
// 		child := item.ChildAt(i)
// 		if child == nil {
// 			continue
// 		}

// 		// 子どものチェック状態を設定
// 		if outputItem, ok := child.(*OutputItem); ok {
// 			outputItem.SetChecked(checked)
// 			treeView.SetChecked(outputItem, checked)
// 		}

// 		// 再帰的に孫も処理（フラグを一時的にクリアして再帰呼び出し）
// 		ss.IsOutputUpdatingChildren = false
// 		ss.SetOutputChildrenChecked(treeView, child, checked)
// 		ss.IsOutputUpdatingChildren = true
// 	}
// }

// // SetOutputPhysicsChecked は物理関連ボーンのチェック状態を設定する
// func (ss *BakeSet) SetOutputPhysicsChecked(treeView *walk.TreeView, item walk.TreeItem, checked bool) {
// 	// if ss.IsOutputUpdatingPhysics {
// 	// 	return
// 	// }

// 	// 無限ループを防ぐためのフラグ
// 	ss.IsOutputUpdatingPhysics = true
// 	defer func() {
// 		ss.IsOutputUpdatingPhysics = false
// 	}()

// 	if item == nil {
// 		for i := range treeView.Model().RootCount() {
// 			item := treeView.Model().RootAt(i)
// 			ss.SetOutputPhysicsChecked(treeView, item, checked)
// 		}
// 		return
// 	}

// 	// 子どもの数を取得
// 	for i := range item.ChildCount() {
// 		child := item.ChildAt(i)
// 		if child == nil {
// 			continue
// 		}

// 		// 出力IKボーンのチェック状態を設定
// 		if outputItem, ok := child.(*OutputItem); ok {
// 			if outputItem.AsPhysics() {
// 				outputItem.SetChecked(checked)
// 				treeView.SetChecked(outputItem, checked)
// 			}
// 		}

// 		// 子どもアイテムのチェック状態を設定
// 		ss.SetOutputPhysicsChecked(treeView, child, checked)
// 	}
// }

// // SetOutputIkChecked はIK関連ボーンのチェック状態を設定する
// func (ss *BakeSet) SetOutputIkChecked(treeView *walk.TreeView, item walk.TreeItem, checked bool) {
// 	// if ss.IsOutputUpdatingIk {
// 	// 	return
// 	// }

// 	if item == nil {
// 		for i := range treeView.Model().RootCount() {
// 			item := treeView.Model().RootAt(i)
// 			ss.SetOutputIkChecked(treeView, item, checked)
// 		}
// 		return
// 	}

// 	// 無限ループを防ぐためのフラグ
// 	ss.IsOutputUpdatingIk = true
// 	defer func() {
// 		ss.IsOutputUpdatingIk = false
// 	}()

// 	// 子どもの数を取得
// 	for i := range item.ChildCount() {
// 		child := item.ChildAt(i)
// 		if child == nil {
// 			continue
// 		}

// 		// 出力IKボーンのチェック状態を設定
// 		if outputItem, ok := child.(*OutputItem); ok {
// 			if outputItem.AsIk() {
// 				outputItem.SetChecked(checked)
// 				treeView.SetChecked(outputItem, checked)
// 			}
// 		}

// 		// 子どもアイテムのチェック状態を設定
// 		ss.SetOutputIkChecked(treeView, child, checked)
// 	}
// }
