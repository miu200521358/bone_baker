package domain

// PhysicsBoneServiceFactory 物理ボーンサービスの生成を担うファクトリ
type PhysicsBoneServiceFactory struct{}

// NewPhysicsBoneServiceFactory ファクトリのコンストラクタ
func NewPhysicsBoneServiceFactory() *PhysicsBoneServiceFactory {
	return &PhysicsBoneServiceFactory{}
}

// CreatePhysicsBoneManager PhysicsBoneManagerインターフェースの実装を作成
func (f *PhysicsBoneServiceFactory) CreatePhysicsBoneManager() PhysicsBoneManager {
	return NewPhysicsBoneService()
}

// CreatePhysicsBoneProcessor PhysicsBoneProcessorインターフェースの実装を作成
func (f *PhysicsBoneServiceFactory) CreatePhysicsBoneProcessor() PhysicsBoneProcessor {
	return NewPhysicsBoneService()
}

// CreatePhysicsBoneDisplayManager PhysicsBoneDisplayManagerインターフェースの実装を作成
func (f *PhysicsBoneServiceFactory) CreatePhysicsBoneDisplayManager() PhysicsBoneDisplayManager {
	return NewPhysicsBoneService()
}

// CreatePhysicsBoneNamer PhysicsBoneNamerインターフェースの実装を作成
func (f *PhysicsBoneServiceFactory) CreatePhysicsBoneNamer() PhysicsBoneNamer {
	return NewPhysicsBoneService()
}

// CreatePhysicsRigidBodyFixer PhysicsRigidBodyFixerインターフェースの実装を作成
func (f *PhysicsBoneServiceFactory) CreatePhysicsRigidBodyFixer() PhysicsRigidBodyFixer {
	return NewPhysicsBoneService()
}

// BoneNameEncodingServiceFactory ボーン名エンコーディングサービスのファクトリ
type BoneNameEncodingServiceFactory struct{}

// NewBoneNameEncodingServiceFactory ファクトリのコンストラクタ
func NewBoneNameEncodingServiceFactory() *BoneNameEncodingServiceFactory {
	return &BoneNameEncodingServiceFactory{}
}

// CreateBoneNameEncodingService BoneNameEncodingServiceの実装を作成
func (f *BoneNameEncodingServiceFactory) CreateBoneNameEncodingService() *BoneNameEncodingService {
	return NewBoneNameEncodingService()
}

// ValueObjectFactory Value Objectsの生成を担うファクトリ
type ValueObjectFactory struct{}

// NewValueObjectFactory ファクトリのコンストラクタ
func NewValueObjectFactory() *ValueObjectFactory {
	return &ValueObjectFactory{}
}

// CreateFilePath FilePathオブジェクトを安全に作成
func (f *ValueObjectFactory) CreateFilePath(path string) *FilePath {
	return NewFilePath(path)
}

// CreateFrameRange FrameRangeオブジェクトを安全に作成
func (f *ValueObjectFactory) CreateFrameRange(startFrame, endFrame float32) (*FrameRange, error) {
	return NewFrameRange(startFrame, endFrame)
}
