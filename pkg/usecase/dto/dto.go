package dto

// PhysicsItemDTO 物理アイテムのデータ転送オブジェクト
type PhysicsItemDTO struct {
	ID             string  `json:"id"`
	Text           string  `json:"text"`
	MassRatio      float64 `json:"mass_ratio"`
	StiffnessRatio float64 `json:"stiffness_ratio"`
	TensionRatio   float64 `json:"tension_ratio"`
	IsSelected     bool    `json:"is_selected"`
	HasPhysics     bool    `json:"has_physics"`
}

// OutputItemDTO 出力アイテムのデータ転送オブジェクト
type OutputItemDTO struct {
	ID        string          `json:"id"`
	Text      string          `json:"text"`
	Checked   bool            `json:"checked"`
	IsIK      bool            `json:"is_ik"`
	IsPhysics bool            `json:"is_physics"`
	Children  []OutputItemDTO `json:"children"`
}

// PhysicsTreeDTO 物理ツリーのデータ転送オブジェクト
type PhysicsTreeDTO struct {
	Items []PhysicsItemDTO `json:"items"`
}

// OutputTreeDTO 出力ツリーのデータ転送オブジェクト
type OutputTreeDTO struct {
	Items []OutputItemDTO `json:"items"`
}

// BakeSetInfoDTO BakeSet情報のデータ転送オブジェクト
type BakeSetInfoDTO struct {
	Index              int     `json:"index"`
	OriginalModelPath  string  `json:"original_model_path"`
	OriginalMotionPath string  `json:"original_motion_path"`
	OutputModelPath    string  `json:"output_model_path"`
	OutputMotionPath   string  `json:"output_motion_path"`
	MaxFrame           float32 `json:"max_frame"`
}
