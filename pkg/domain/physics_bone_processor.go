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

// PhysicsBoneManager 物理ボーン管理の複合インターフェース
// 必要に応じて複数の機能を組み合わせて使用
type PhysicsBoneManager interface {
	PhysicsBoneProcessor
	PhysicsBoneDisplayManager
	PhysicsBoneNamer
	PhysicsRigidBodyFixer
}
