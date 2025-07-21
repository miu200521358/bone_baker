package domain

import (
	"fmt"
	"math"
	"sync"

	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/mlib_go/pkg/domain/vmd"
	"github.com/miu200521358/mlib_go/pkg/infrastructure/mfile"
	"github.com/miu200521358/mlib_go/pkg/infrastructure/repository"
)

type PhysicsSet struct {
	Index       int  // インデックス
	IsTerminate bool // 処理停止フラグ

	OriginalMotionPath string `json:"original_motion_path"` // 元モーションパス
	OriginalModelPath  string `json:"original_model_path"`  // 元モデルパス
	OutputMotionPath   string `json:"-"`                    // 出力モーションパス
	OutputModelPath    string `json:"-"`                    // 出力モデルパス

	OriginalMotionName string `json:"-"` // 元モーション名
	OriginalModelName  string `json:"-"` // 元モーション名
	OutputModelName    string `json:"-"` // 物理焼き込み先モデル名

	OriginalMotion    *vmd.VmdMotion `json:"-"` // 元モデル
	OriginalModel     *pmx.PmxModel  `json:"-"` // 元モデル
	PhysicsBakedModel *pmx.PmxModel  `json:"-"` // 物理焼き込み先モデル
	OutputMotion      *vmd.VmdMotion `json:"-"` // 出力結果モーション
}

func NewPhysicsSet(index int) *PhysicsSet {
	return &PhysicsSet{
		Index: index,
	}
}

func (ss *PhysicsSet) CreateOutputModelPath() string {
	if ss.OriginalModel == nil {
		return ""
	}

	// 物理焼き込み先モデルが指定されている場合、ファイル名を含める
	return mfile.CreateOutputPath(ss.OriginalModel.Path(), "PF")
}

func (ss *PhysicsSet) CreateOutputMotionPath() string {
	if ss.OriginalMotion == nil || ss.PhysicsBakedModel == nil {
		return ""
	}

	// 物理焼き込み先モデルが指定されている場合、ファイル名を含める
	_, fileName, _ := mfile.SplitPath(ss.PhysicsBakedModel.Path())

	return mfile.CreateOutputPath(
		ss.OriginalMotion.Path(), fmt.Sprintf("PF_%s", fileName))
}

func (ss *PhysicsSet) setMotion(originalMotion, outputMotion *vmd.VmdMotion) {
	if originalMotion == nil || outputMotion == nil {
		ss.OriginalMotionPath = ""
		ss.OriginalMotionName = ""
		ss.OriginalMotion = nil

		ss.OutputMotionPath = ""
		ss.OutputMotion = vmd.NewVmdMotion("")
		ss.OutputMotion.BoneFrames.SetDisablePhysics(true) // 物理演算無効をONにする

		return
	}

	ss.OriginalMotionName = originalMotion.Name()
	ss.OriginalMotion = originalMotion
	ss.OutputMotion = outputMotion
}

func (ss *PhysicsSet) setModels(originalModel, physicsBakedModel *pmx.PmxModel) {
	if originalModel == nil {
		ss.OriginalModelPath = ""
		ss.OriginalModelName = ""
		ss.OriginalModel = nil
		ss.PhysicsBakedModel = nil
		return
	}

	ss.OriginalModelPath = originalModel.Path()
	ss.OriginalModelName = originalModel.Name()
	ss.OriginalModel = originalModel
	ss.PhysicsBakedModel = physicsBakedModel
}

// LoadModel モデルを読み込む
func (ss *PhysicsSet) LoadModel(path string) error {
	if path == "" {
		ss.setModels(nil, nil)
		return nil
	}

	var wg sync.WaitGroup
	var originalModel, physicsBakedModel *pmx.PmxModel

	errChan := make(chan error, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()

		pmxRep := repository.NewPmxRepository(true)
		if data, err := pmxRep.Load(path); err == nil {
			originalModel = data.(*pmx.PmxModel)
			if err := originalModel.Bones.InsertShortageOverrideBones(); err != nil {
				mlog.ET(mi18n.T("システム用ボーン追加失敗"), err, "")
				errChan <- err
			} else {
				// 物理ボーンの名前に接頭辞を追加
				ss.insertPhysicsBonePrefix(originalModel)
			}
		} else {
			mlog.ET(mi18n.T("読み込み失敗"), err, "")
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		pmxRep := repository.NewPmxRepository(false)
		if data, err := pmxRep.Load(path); err == nil {
			physicsBakedModel = data.(*pmx.PmxModel)
			if err := physicsBakedModel.Bones.InsertShortageOverrideBones(); err != nil {
				mlog.ET(mi18n.T("システム用ボーン追加失敗"), err, "")
				errChan <- err
			} else {
				// 物理ボーンの名前に接頭辞を追加
				ss.insertPhysicsBonePrefix(physicsBakedModel)
				// 物理ボーンの物理剛体を無効化
				ss.fixPhysicsRigidBodies(physicsBakedModel)
			}
		} else {
			mlog.ET(mi18n.T("読み込み失敗"), err, "")
			errChan <- err
		}
	}()

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			ss.setModels(nil, nil)
			return err
		}
	}

	ss.setModels(originalModel, physicsBakedModel)
	ss.OutputModelPath = ss.CreateOutputModelPath()

	return nil
}

func (ss *PhysicsSet) insertPhysicsBonePrefix(model *pmx.PmxModel) {
	if model == nil {
		return
	}

	digits := int(math.Log10(float64(model.Bones.Length()))) + 1

	// 物理ボーンの名前に接頭辞を追加
	model.Bones.ForEach(func(boneIndex int, bone *pmx.Bone) bool {
		if bone.RigidBody != nil && bone.RigidBody.PhysicsType != pmx.PHYSICS_TYPE_STATIC {
			// ボーンINDEXを0埋めして設定
			formattedBoneName := fmt.Sprintf("PF%0*d_%s", digits, boneIndex, bone.Name())
			bone.SetName(formattedBoneName)
		}
		return true
	})
}

func (ss *PhysicsSet) fixPhysicsRigidBodies(model *pmx.PmxModel) {
	if model == nil {
		return
	}

	// 物理ボーンの剛体を修正
	model.RigidBodies.ForEach(func(rigidBodyIndex int, rigidBody *pmx.RigidBody) bool {
		rigidBody.PhysicsType = pmx.PHYSICS_TYPE_STATIC // 剛体の物理演算を無効にする
		model.RigidBodies.Update(rigidBody)
		return true
	})
}

func (ss *PhysicsSet) LoadMotion(path string) error {

	if path == "" {
		ss.setMotion(nil, nil)

		return nil
	}

	var wg sync.WaitGroup
	var originalMotion, outputMotion *vmd.VmdMotion
	errChan := make(chan error, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()

		vmdRep := repository.NewVmdVpdRepository(false)
		if data, err := vmdRep.Load(path); err == nil {
			originalMotion = data.(*vmd.VmdMotion)
		} else {
			mlog.ET(mi18n.T("読み込み失敗"), err, "")
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		vmdRep := repository.NewVmdVpdRepository(true)
		if data, err := vmdRep.Load(path); err == nil {
			outputMotion = data.(*vmd.VmdMotion)
			outputMotion.BoneFrames.SetDisablePhysics(true) // 物理演算無効をONにする
		} else {
			mlog.ET(mi18n.T("読み込み失敗"), err, "")
			errChan <- err
		}
	}()

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	ss.setMotion(originalMotion, outputMotion)

	// 出力パスを設定
	ss.OutputMotionPath = ss.CreateOutputMotionPath()

	return nil
}

func (ss *PhysicsSet) Delete() {
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
func (ss *PhysicsSet) GetOutputMotionOnlyPhysics() *vmd.VmdMotion {
	if ss.OriginalModel == nil || ss.OutputMotion == nil {
		return nil
	}

	motion := vmd.NewVmdMotion(ss.OutputMotionPath)

	// 物理無効ON
	motion.BoneFrames.SetDisablePhysics(true)

	ss.OutputMotion.BoneFrames.ForEach(func(boneName string, boneNameFrames *vmd.BoneNameFrames) {
		if bone, err := ss.OriginalModel.Bones.GetByName(boneName); err == nil {
			if bone.RigidBody != nil && bone.RigidBody.PhysicsType != pmx.PHYSICS_TYPE_STATIC {
				// 物理剛体がくっついているボーンのみ登録対象
				motion.BoneFrames.Update(boneNameFrames)
			}
		}
	})

	return motion
}
