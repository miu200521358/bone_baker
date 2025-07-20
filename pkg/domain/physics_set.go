package domain

import (
	"fmt"
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
	PhysicsModelPath   string `json:"physics_model_path"`   // 物理焼き込み先モデルパス
	OutputMotionPath   string `json:"-"`                    // 出力モーションパス
	OutputModelPath    string `json:"-"`                    // 出力モデルパス

	OriginalMotionName string `json:"-"` // 元モーション名
	OriginalModelName  string `json:"-"` // 元モーション名
	OutputModelName    string `json:"-"` // 物理焼き込み先モデル名

	OriginalMotion      *vmd.VmdMotion `json:"-"` // 元モデル
	OriginalModel       *pmx.PmxModel  `json:"-"` // 元モデル
	OriginalConfigModel *pmx.PmxModel  `json:"-"` // 元モデル(ボーン追加)
	PhysicsModel        *pmx.PmxModel  `json:"-"` // 物理焼き込み先モデル
	PhysicsConfigModel  *pmx.PmxModel  `json:"-"` // 物理焼き込み先モデル(ボーン追加)
	OutputMotion        *vmd.VmdMotion `json:"-"` // 出力結果モーション

}

func NewPhysicsSet(index int) *PhysicsSet {
	return &PhysicsSet{
		Index: index,
	}
}

func (ss *PhysicsSet) CreateOutputModelPath() string {
	if ss.PhysicsModelPath == "" {
		return ""
	}
	// 物理焼き込み先モデルが指定されている場合、ファイル名を含める
	_, fileName, _ := mfile.SplitPath(ss.PhysicsModelPath)

	return mfile.CreateOutputPath(ss.PhysicsModelPath, fileName)
}

func (ss *PhysicsSet) CreateOutputMotionPath() string {
	if ss.OriginalMotionPath == "" || ss.PhysicsModelPath == "" {
		return ""
	}

	// 物理焼き込み先モデルが指定されている場合、ファイル名を含める
	_, fileName, _ := mfile.SplitPath(ss.PhysicsModelPath)

	return mfile.CreateOutputPath(
		ss.OriginalMotionPath, fmt.Sprintf("%s%s%02d", fileName, "PF", ss.Index))
}

func (ss *PhysicsSet) setMotion(originalMotion, outputMotion *vmd.VmdMotion) {
	if originalMotion == nil || outputMotion == nil {
		ss.OriginalMotionPath = ""
		ss.OriginalMotionName = ""
		ss.OriginalMotion = nil

		ss.OutputMotionPath = ""
		ss.OutputMotion = nil

		return
	}

	ss.OriginalMotionPath = originalMotion.Path()
	ss.OriginalMotionName = originalMotion.Name()
	ss.OriginalMotion = originalMotion

	ss.OutputMotionPath = outputMotion.Path()
	ss.OutputMotion = outputMotion

}

func (ss *PhysicsSet) setOriginalModel(originalModel, originalConfigModel *pmx.PmxModel) {
	if originalModel == nil {
		ss.OriginalModelPath = ""
		ss.OriginalModelName = ""
		ss.OriginalModel = nil
		ss.OriginalConfigModel = nil
		return
	}

	ss.OriginalModelPath = originalModel.Path()
	ss.OriginalModelName = originalModel.Name()
	ss.OriginalModel = originalModel
	ss.OriginalConfigModel = originalConfigModel
}

func (ss *PhysicsSet) setPhysicsModel(physicsModel, physicsConfigModel *pmx.PmxModel) {
	if physicsModel == nil || physicsConfigModel == nil {
		ss.PhysicsModelPath = ""
		ss.OutputModelName = ""
		ss.PhysicsConfigModel = nil
		ss.PhysicsModel = nil
		return
	}

	ss.PhysicsModelPath = physicsModel.Path()
	ss.OutputModelName = physicsModel.Name()
	ss.PhysicsModel = physicsModel
	ss.PhysicsConfigModel = physicsConfigModel
}

// LoadOriginalModel 物理焼き込み元モデルを読み込む
// TODO json の場合はフィッティングあり
func (ss *PhysicsSet) LoadOriginalModel(path string) error {
	if path == "" {
		ss.setOriginalModel(nil, nil)
		return nil
	}

	var wg sync.WaitGroup
	var originalModel, originalConfigModel *pmx.PmxModel

	wg.Add(1)
	errChan := make(chan error, 2)
	go func() {
		defer wg.Done()

		pmxRep := repository.NewPmxRepository(true)
		if data, err := pmxRep.Load(path); err == nil {
			originalModel = data.(*pmx.PmxModel)

			if err := originalModel.Bones.InsertShortageOverrideBones(); err != nil {
				mlog.ET(mi18n.T("システム用ボーン追加失敗"), err, "")
				errChan <- err
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
			originalConfigModel = data.(*pmx.PmxModel)
			if err := originalConfigModel.Bones.InsertShortageOverrideBones(); err != nil {
				mlog.ET(mi18n.T("システム用ボーン追加失敗"), err, "")
				errChan <- err
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
			ss.setOriginalModel(nil, nil)
			return err
		}
	}

	// 元モデル設定
	ss.setOriginalModel(originalModel, originalConfigModel)

	// 出力パスを設定
	ss.OutputModelPath = ss.CreateOutputModelPath()

	return nil
}

// LoadPhysicsModel 物理焼き込み先モデルを読み込む
func (ss *PhysicsSet) LoadPhysicsModel(path string) error {
	if path == "" {
		ss.setPhysicsModel(nil, nil)
		return nil
	}

	var wg sync.WaitGroup
	var physicsModel, physicsConfigModel *pmx.PmxModel

	errChan := make(chan error, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()

		pmxRep := repository.NewPmxRepository(true)
		if data, err := pmxRep.Load(path); err == nil {
			physicsModel = data.(*pmx.PmxModel)
			if err := physicsModel.Bones.InsertShortageOverrideBones(); err != nil {
				mlog.ET(mi18n.T("システム用ボーン追加失敗"), err, "")
				errChan <- err
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
			physicsConfigModel = data.(*pmx.PmxModel)
			if err := physicsConfigModel.Bones.InsertShortageOverrideBones(); err != nil {
				mlog.ET(mi18n.T("システム用ボーン追加失敗"), err, "")
				errChan <- err
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
			ss.setPhysicsModel(nil, nil)
			return err
		}
	}

	// 物理焼き込みモデル設定
	ss.setPhysicsModel(physicsModel, physicsConfigModel)

	// 出力パスを設定
	ss.OutputModelPath = ss.CreateOutputModelPath()
	ss.OutputMotionPath = ss.CreateOutputMotionPath()

	return nil
}

// LoadMotion 物理焼き込み対象モーションを読み込む
func (ss *PhysicsSet) LoadMotion(path string) error {

	if path == "" {
		ss.setMotion(nil, nil)
		return nil
	}

	var wg sync.WaitGroup
	var originalMotion, physicsMotion *vmd.VmdMotion
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
			physicsMotion = data.(*vmd.VmdMotion)
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

	ss.setMotion(originalMotion, physicsMotion)

	// 出力パスを設定
	outputPath := ss.CreateOutputMotionPath()
	ss.OutputMotionPath = outputPath

	return nil
}

func (ss *PhysicsSet) Delete() {
	ss.OriginalMotionPath = ""
	ss.OriginalModelPath = ""
	ss.PhysicsModelPath = ""
	ss.OutputMotionPath = ""
	ss.OutputModelPath = ""

	ss.OriginalMotionName = ""
	ss.OriginalModelName = ""
	ss.OutputModelName = ""

	ss.OriginalMotion = nil
	ss.OriginalModel = nil
	ss.OriginalConfigModel = nil
	ss.PhysicsModel = nil
	ss.PhysicsConfigModel = nil
	ss.OutputMotion = nil
}
