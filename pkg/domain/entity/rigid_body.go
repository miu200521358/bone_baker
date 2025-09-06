package entity

import (
	"strings"

	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/domain/mmath"
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
)

type RigidBodyRecord struct {
	StartFrame    float32        `json:"start_frame"`     // 区間開始フレーム
	EndFrame      float32        `json:"end_frame"`       // 区間終了フレーム
	MaxStartFrame float32        `json:"max_start_frame"` // 最大値開始フレーム
	MaxEndFrame   float32        `json:"max_end_frame"`   // 最大値終了フレーム
	Tree          *RigidBodyTree `json:"items"`           // 剛体アイテム一覧
}

func NewRigidBodyRecord(startFrame, endFrame float32, model *pmx.PmxModel) *RigidBodyRecord {
	return &RigidBodyRecord{
		StartFrame:    startFrame,
		MaxStartFrame: startFrame,
		MaxEndFrame:   endFrame,
		EndFrame:      endFrame,
		Tree:          newRigidBodyTree(model),
	}
}

func (r *RigidBodyRecord) ItemNames() string {
	var names []string
	for _, item := range r.Tree.Items {
		names = append(names, item.ItemNames(names)...)
	}

	if len(names) == 0 {
		return mi18n.T("変更剛体なし")
	}

	return strings.Join(names, ", ")
}

type RigidBodyTree struct {
	Items []*RigidBodyItem
}

func newRigidBodyTree(model *pmx.PmxModel) *RigidBodyTree {
	items := &RigidBodyTree{}

	for _, boneIndex := range model.Bones.LayerSortedIndexes {
		if bone, err := model.Bones.Get(boneIndex); err == nil {
			parent := items.AtByBoneIndex(bone.ParentIndex)
			if len(bone.RigidBodies) == 0 {
				// 自身に剛体が無い場合、そのまま剛体なしで追加
				item := newRigidBodyItem(bone, nil, parent)
				if parent == nil {
					items.AddNode(item)
				} else {
					parent.AddChild(item)
				}
				continue
			}

			for _, rigidBody := range bone.RigidBodies {
				item := newRigidBodyItem(bone, rigidBody, parent)
				if parent == nil {
					items.AddNode(item)
				} else {
					parent.AddChild(item)
				}
			}
		}
	}

	// 剛体を持つボーンのみを保存
	items.SaveOnlyRigidBodyItems()

	return items
}

// 物理ボーンを含むツリーだけ残す
func (r *RigidBodyTree) SaveOnlyRigidBodyItems() {
	newNodes := make([]*RigidBodyItem, 0)
	for _, node := range r.Items {
		// 子に物理ボーンがある場合のみ残す
		node.saveOnlyRigidBodyItems()

		if node.hasRigidBodyChild() {
			newNodes = append(newNodes, node)
		}
	}

	r.Items = newNodes
}

func (r *RigidBodyTree) AddNode(item *RigidBodyItem) {
	r.Items = append(r.Items, item)
}

func (r RigidBodyTree) AtByBoneIndex(boneIndex int) *RigidBodyItem {
	if boneIndex < 0 {
		return nil
	}

	for _, item := range r.Items {
		if found := item.AtByBoneIndex(boneIndex); found != nil {
			return found
		}
	}

	return nil
}

type RigidBodyItem struct {
	Bone           *pmx.Bone        // 剛体に紐付くボーン情報
	RigidBody      *pmx.RigidBody   // 剛体情報
	SizeRatio      *mmath.MVec3     `json:"size_ratio"`       // 大きさ比率
	MassRatio      float64          `json:"mass_ratio"`       // 質量比率
	StiffnessRatio float64          `json:"stiffness_ratio"`  // 硬さ比率
	TensionRatio   float64          `json:"tension_ratio"`    // 張り比率
	Modified       bool             `json:"modified"`         // 変更されたかどうか
	RigidBodyIndex int              `json:"rigid_body_index"` // 剛体インデックス
	RigidBodyName  string           `json:"rigid_body_name"`  // 剛体名
	Parent         *RigidBodyItem   `json:"parent"`           // 親剛体アイテム
	Children       []*RigidBodyItem `json:"children"`         // 子剛体アイテム
}

func newRigidBodyItem(bone *pmx.Bone, rigidBody *pmx.RigidBody, parent *RigidBodyItem) *RigidBodyItem {
	item := &RigidBodyItem{
		Bone:           bone,
		RigidBody:      rigidBody,
		SizeRatio:      &mmath.MVec3{X: 1.0, Y: 1.0, Z: 1.0},
		MassRatio:      1,
		StiffnessRatio: 1,
		TensionRatio:   1,
		Modified:       false,
		Parent:         parent,
		Children:       []*RigidBodyItem{},
	}

	if rigidBody != nil {
		item.RigidBodyIndex = rigidBody.Index()
		item.RigidBodyName = rigidBody.Name()
	}

	return item
}

func (pi *RigidBodyItem) ItemNames(names []string) []string {
	if pi.RigidBody != nil && pi.Modified {
		names = append(names, pi.RigidBody.Name())
	}

	for _, child := range pi.Children {
		names = child.ItemNames(names)
	}

	return names
}

func (pi *RigidBodyItem) AtByBoneIndex(boneIndex int) *RigidBodyItem {
	if pi.Bone == nil {
		return nil
	}

	if pi.Bone.Index() == boneIndex {
		return pi
	}

	for _, child := range pi.Children {
		if found := child.AtByBoneIndex(boneIndex); found != nil {
			return found
		}
	}

	return nil
}

func (pi *RigidBodyItem) AddChild(child *RigidBodyItem) {
	pi.Children = append(pi.Children, child)
}

func (pi *RigidBodyItem) saveOnlyRigidBodyItems() {
	newChildren := make([]*RigidBodyItem, 0)
	for _, child := range pi.Children {
		child.saveOnlyRigidBodyItems()

		if child.hasRigidBodyChild() {
			newChildren = append(newChildren, child)
		}
	}

	pi.Children = newChildren
}

func (pi *RigidBodyItem) hasRigidBodyChild() bool {
	hasRigidBody := false
	for _, c := range pi.Children {
		if c.hasRigidBodyChild() {
			hasRigidBody = true
			break
		}
	}

	return hasRigidBody || pi.RigidBody != nil
}
