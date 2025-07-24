package domain

import (
	"errors"
	"fmt"

	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
	"github.com/miu200521358/mlib_go/pkg/infrastructure/mfile"
)

type BakeSet struct {
	Index       int  // インデックス
	IsTerminate bool // 処理停止フラグ

	OriginalMotionPath string `json:"original_motion_path"` // 元モーションパス
	OriginalModelPath  string `json:"original_model_path"`  // 元モデルパス
	OutputMotionPath   string `json:"-"`                    // 出力モーションパス
	OutputModelPath    string `json:"-"`                    // 出力モデルパス

	OriginalMotionName string `json:"-"` // 元モーション名
	OriginalModelName  string `json:"-"` // 元モーション名
	OutputModelName    string `json:"-"` // 物理焼き込み先モデル名

	OriginalMotion *vmd.VmdMotion `json:"-"` // 元モデル
	OriginalModel  *pmx.PmxModel  `json:"-"` // 元モデル
	BakedModel     *pmx.PmxModel  `json:"-"` // 物理焼き込み先モデル
	OutputMotion   *vmd.VmdMotion `json:"-"` // 出力結果モーション

	PhysicsTree *PhysicsModel `json:"-"` // 物理ボーンツリー
	OutputTree  *OutputModel  `json:"-"` // 出力ボーンツリー
}

func NewPhysicsSet(index int) *BakeSet {
	return &BakeSet{
		Index:       index,
		PhysicsTree: NewPhysicsModel(),
	}
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
		ss.OriginalMotionPath = ""
		ss.OriginalMotionName = ""
		ss.OriginalMotion = nil

		ss.OutputMotionPath = ""
		ss.OutputMotion = vmd.NewVmdMotion("")

		return
	}

	ss.OriginalMotionName = originalMotion.Name()
	ss.OriginalMotion = originalMotion
	ss.OutputMotion = outputMotion
}

func (ss *BakeSet) setModels(originalModel, physicsBakedModel *pmx.PmxModel) {
	if originalModel == nil {
		ss.OriginalModelPath = ""
		ss.OriginalModelName = ""
		ss.OriginalModel = nil
		ss.BakedModel = nil
		return
	}

	ss.OriginalModelPath = originalModel.Path()
	ss.OriginalModelName = originalModel.Name()
	ss.OriginalModel = originalModel
	ss.BakedModel = physicsBakedModel
}

// SetModels モデルを設定（ドメインサービスを使用）
func (ss *BakeSet) SetModels(originalModel, bakedModel *pmx.PmxModel) error {
	if originalModel == nil {
		ss.setModels(nil, nil)
		return nil
	}

	ss.setModels(originalModel, bakedModel)
	ss.OutputModelPath = ss.CreateOutputModelPath()

	return nil
}

// ClearModels モデルをクリア（公開メソッド）
func (ss *BakeSet) ClearModels() {
	ss.setModels(nil, nil)
}

// SetMotions モーションを設定（公開メソッド）
func (ss *BakeSet) SetMotions(originalMotion, outputMotion *vmd.VmdMotion) error {
	ss.setMotion(originalMotion, outputMotion)
	ss.OutputMotionPath = ss.CreateOutputMotionPath()
	return nil
}

// ClearMotions モーションをクリア（公開メソッド）
func (ss *BakeSet) ClearMotions() {
	ss.setMotion(nil, nil)
}

func (ss *BakeSet) Delete() {
	ss.OriginalMotionPath = ""
	ss.OriginalModelPath = ""
	ss.OutputMotionPath = ""
	ss.OutputModelPath = ""

	ss.OriginalMotionName = ""
	ss.OriginalModelName = ""
	ss.OutputModelName = ""

	ss.OriginalMotion = nil
	ss.OriginalModel = nil
	ss.OutputMotion = nil
}

// 物理ボーンだけ残す
func (ss *BakeSet) GetOutputMotionOnlyChecked(startFrame, endFrame float64) ([]*vmd.VmdMotion, error) {
	motions := make([]*vmd.VmdMotion, 0)

	if ss.OriginalModel == nil || ss.OutputMotion == nil {
		return nil, errors.New(mi18n.T("物理焼き込みセットの元モデルまたは出力モーションが設定されていません"))
	}

	if startFrame < 0 || endFrame < 0 || startFrame > endFrame {
		return nil, errors.New(mi18n.T("開始フレームより終了フレームが小さいか、負の値が設定されています"))
	}

	boneCount := 0
	ss.OriginalModel.Bones.ForEach(func(boneIndex int, bone *pmx.Bone) bool {
		item := ss.OutputTree.AtByBoneIndex(boneIndex)
		if item == nil || !item.(*OutputItem).Checked() {
			// チェックされていないボーンはスキップ
			return true
		}

		boneCount++
		return true
	})

	nextFrameCount := 0
	logFrameCount := 0

	motion := vmd.NewVmdMotion(ss.OutputMotionPath)
	dirPath, fileName, ext := mfile.SplitPath(ss.OutputMotionPath)
	motion = vmd.NewVmdMotion(fmt.Sprintf("%s/%s_%04.0f%s", dirPath, fileName, startFrame, ext))

	// ボーン焼き込み
	for index := startFrame; index <= endFrame; index++ {
		nextFrameCount += boneCount

		if nextFrameCount > vmd.MAX_BONE_FRAMES {
			// キーフレーム数が上限を超える場合は切り替える
			motions = append(motions, motion)

			mlog.I(fmt.Sprintf(mi18n.T("キーフレーム数が上限を超えるため、モーションを切り替えます[%04.0fF]: %d -> %d"),
				index, nextFrameCount, vmd.MAX_BONE_FRAMES))

			dirPath, fileName, ext := mfile.SplitPath(ss.OutputMotionPath)
			motion = vmd.NewVmdMotion(fmt.Sprintf("%s/%s_%04.0f%s", dirPath, fileName, index, ext))
			nextFrameCount = boneCount
			logFrameCount = boneCount
		}

		if nextFrameCount/100000 > logFrameCount/100000 {
			mlog.I(fmt.Sprintf(mi18n.T("- 物理焼き込み中... [%04.0fF] %dキーフレーム"), index, nextFrameCount))
			logFrameCount = nextFrameCount
		}

		ss.OriginalModel.Bones.ForEach(func(boneIndex int, bone *pmx.Bone) bool {
			item := ss.OutputTree.AtByBoneIndex(boneIndex)
			if item == nil || !item.(*OutputItem).Checked() {
				// チェックされていないボーンはスキップ
				return true
			}

			bf := ss.OutputMotion.BoneFrames.Get(bone.Name()).Get(float32(index))
			if bf == nil {
				return true
			}

			if bone.HasPhysics() {
				bf.DisablePhysics = true // 物理演算を無効にする
			}
			motion.AppendBoneFrame(bone.Name(), bf)

			return true
		})
	}

	// 最後に物理演算を有効にする
	ss.OriginalModel.Bones.ForEach(func(boneIndex int, bone *pmx.Bone) bool {
		item := ss.OutputTree.AtByBoneIndex(boneIndex)
		if item == nil || !item.(*OutputItem).Checked() {
			// チェックされていないボーンはスキップ
			return true
		}

		if bone.HasPhysics() {
			// 最後に物理有効化を入れる
			lastBf := ss.OutputMotion.BoneFrames.Get(bone.Name()).Get(float32(endFrame + 1))
			if lastBf == nil {
				return true
			}
			lastBf.DisablePhysics = false // 物理演算を有効にする
			motion.AppendBoneFrame(bone.Name(), lastBf)
		}

		return true
	})

	motions = append(motions, motion)
	return motions, nil
}
