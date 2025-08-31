package domain

import "github.com/miu200521358/mlib_go/pkg/domain/pmx"

// PhysicsBoneProcessor 物理ボーン処理の基本インターフェース
type PhysicsBoneProcessor interface {
	ProcessPhysicsBones(model *pmx.PmxModel)
}

// PhysicsBoneDisplayManager 物理ボーンの表示管理インターフェース
type PhysicsBoneDisplayManager interface {
	AppendPhysicsBoneToDisplaySlots(model *pmx.PmxModel)
}

// PhysicsBoneNamer 物理ボーンの名前管理インターフェース
type PhysicsBoneNamer interface {
	InsertPhysicsBonePrefix(model *pmx.PmxModel)
}

// PhysicsRigidBodyFixer 物理剛体修正インターフェース
type PhysicsRigidBodyFixer interface {
	FixPhysicsRigidBodies(model *pmx.PmxModel)
}
