package domain

import (
	"fmt"

	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/domain/mmath"
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/walk/pkg/walk"
)

type PhysicsItem struct {
	bones          *pmx.Bones     // ボーン一覧
	bone           *pmx.Bone      // 剛体に紐付くボーン情報
	rigidBody      *pmx.RigidBody // 剛体情報
	parent         walk.TreeItem
	children       []walk.TreeItem
	sizeRatio      *mmath.MVec3 // 大きさ比率
	massRatio      float64      // 質量比率
	stiffnessRatio float64      // 硬さ比率
	tensionRatio   float64      // 張り比率
}

func NewPhysicsItem(bones *pmx.Bones, bone *pmx.Bone, rigidBody *pmx.RigidBody, parent walk.TreeItem) *PhysicsItem {
	return &PhysicsItem{
		bones:          bones,
		bone:           bone,
		rigidBody:      rigidBody,
		parent:         parent,
		children:       make([]walk.TreeItem, 0),
		sizeRatio:      &mmath.MVec3{X: 1.0, Y: 1.0, Z: 1.0},
		massRatio:      1.0,
		stiffnessRatio: 1.0,
		tensionRatio:   1.0,
	}
}

func (pi *PhysicsItem) Text() string {
	if pi.rigidBody == nil {
		return fmt.Sprintf(mi18n.T("%s (剛体なし)"), pi.bone.Name())
	}

	var nameText string
	if pi.bone != nil {
		nameText = pi.bone.Name()
	} else if pi.rigidBody != nil {
		nameText = pi.rigidBody.Name()
	} else {
		nameText = "Unknown"
	}

	var sizeText string
	switch pi.rigidBody.ShapeType {
	case pmx.SHAPE_SPHERE:
		sizeText = fmt.Sprintf(mi18n.T("半径: %.2f"), pi.sizeRatio.X)
	case pmx.SHAPE_BOX:
		sizeText = fmt.Sprintf(mi18n.T("幅: %.2f, 高さ: %.2f, 奥行: %.2f"), pi.sizeRatio.X, pi.sizeRatio.Y, pi.sizeRatio.Z)
	case pmx.SHAPE_CAPSULE:
		sizeText = fmt.Sprintf(mi18n.T("半径: %.2f, 高さ: %.2f"), pi.sizeRatio.X, pi.sizeRatio.Y)
	}

	return fmt.Sprintf(mi18n.T("%s (大きさ: [%s], 質量: %.2f, 硬さ: %.2f, 張り: %.2f)"),
		nameText, sizeText, pi.massRatio, pi.stiffnessRatio, pi.tensionRatio)
}

func (pi *PhysicsItem) Parent() walk.TreeItem {
	if pi.parent == nil {
		return nil
	}
	return pi.parent
}

func (pi *PhysicsItem) AddChild(child walk.TreeItem) {
	pi.children = append(pi.children, child)
}

func (pi *PhysicsItem) HasPhysicsChild() bool {
	if len(pi.children) == 0 {
		return pi.rigidBody != nil
	}

	hasPhysicsRigidBody := false
	for _, c := range pi.children {
		if c.(*PhysicsItem).HasPhysicsChild() {
			hasPhysicsRigidBody = true
			break
		}
	}

	return hasPhysicsRigidBody || pi.rigidBody != nil
}

func (pi *PhysicsItem) Reset() {
	pi.sizeRatio = &mmath.MVec3{X: 1.0, Y: 1.0, Z: 1.0} // 大きさ比率を初期化
	pi.massRatio = 1.0
	pi.stiffnessRatio = 1.0
	pi.tensionRatio = 1.0

	for _, child := range pi.children {
		child.(*PhysicsItem).Reset()
	}
}

func (pi *PhysicsItem) CalcSizeX(x float64) {
	pi.sizeRatio.X = x

	for _, child := range pi.children {
		child.(*PhysicsItem).CalcSizeX(x)
	}
}

func (pi *PhysicsItem) CalcSizeY(y float64) {
	pi.sizeRatio.Y = y

	for _, child := range pi.children {
		child.(*PhysicsItem).CalcSizeY(y)
	}
}

func (pi *PhysicsItem) CalcSizeZ(z float64) {
	pi.sizeRatio.Z = z

	for _, child := range pi.children {
		child.(*PhysicsItem).CalcSizeZ(z)
	}
}

func (pi *PhysicsItem) CalcMass(massRatio float64) {
	pi.massRatio = massRatio

	for _, child := range pi.children {
		child.(*PhysicsItem).CalcMass(massRatio)
	}
}

func (pi *PhysicsItem) CalcStiffness(stiffnessRatio float64) {
	pi.stiffnessRatio = stiffnessRatio

	for _, child := range pi.children {
		child.(*PhysicsItem).CalcStiffness(stiffnessRatio)
	}
}

func (pi *PhysicsItem) CalcTension(tensionRatio float64) {
	pi.tensionRatio = tensionRatio

	for _, child := range pi.children {
		child.(*PhysicsItem).CalcTension(tensionRatio)
	}
}

func (pi *PhysicsItem) SizeRatio() *mmath.MVec3 {
	return pi.sizeRatio
}

func (pi *PhysicsItem) MassRatio() float64 {
	return pi.massRatio
}

func (pi *PhysicsItem) StiffnessRatio() float64 {
	return pi.stiffnessRatio
}

func (pi *PhysicsItem) TensionRatio() float64 {
	return pi.tensionRatio
}

func (pi *PhysicsItem) SaveOnlyPhysicsItems() {
	newChildren := make([]walk.TreeItem, 0)
	for _, child := range pi.children {
		child.(*PhysicsItem).SaveOnlyPhysicsItems()

		if child.(*PhysicsItem).HasPhysicsChild() {
			newChildren = append(newChildren, child)
		}
	}

	pi.children = newChildren
}

func (pi *PhysicsItem) ChildCount() int {
	return len(pi.children)
}

func (pi *PhysicsItem) HasChild() bool {
	return len(pi.children) > 0
}

func (pi *PhysicsItem) ChildAt(index int) walk.TreeItem {
	if index < 0 || index >= len(pi.children) {
		return nil
	}
	return pi.children[index]
}

func (pi *PhysicsItem) AtByBoneIndex(boneIndex int) *PhysicsItem {
	if pi.bone == nil {
		return nil
	}

	if pi.bone != nil && pi.bone.Index() == boneIndex {
		return pi
	}

	for _, child := range pi.children {
		if found := child.(*PhysicsItem).AtByBoneIndex(boneIndex); found != nil {
			return found
		}
	}

	return nil
}

type PhysicsRigidBodyTreeModel struct {
	*walk.TreeModelBase
	nodes []*PhysicsItem
}

func NewPhysicsRigidBodyTreeModel(model *pmx.PmxModel) *PhysicsRigidBodyTreeModel {
	tree := &PhysicsRigidBodyTreeModel{
		TreeModelBase: &walk.TreeModelBase{},
		nodes:         make([]*PhysicsItem, 0),
	}

	registeredRigidBodyIndexes := make([]bool, model.RigidBodies.Length())

	for _, boneIndex := range model.Bones.LayerSortedIndexes {
		if bone, err := model.Bones.Get(boneIndex); err == nil {
			parent := tree.AtByBoneIndex(bone.ParentIndex)
			if len(bone.RigidBodies) == 0 {
				// 自身に剛体が無い場合、そのまま剛体なしで追加
				item := NewPhysicsItem(model.Bones, bone, nil, parent)
				if parent == nil {
					tree.AddNode(item)
				} else {
					parent.(*PhysicsItem).AddChild(item)
				}
				continue
			}

			for _, rigidBody := range bone.RigidBodies {
				item := NewPhysicsItem(model.Bones, bone, rigidBody, parent)
				if parent == nil {
					tree.AddNode(item)
				} else {
					parent.(*PhysicsItem).AddChild(item)
				}
				registeredRigidBodyIndexes[rigidBody.Index()] = true
			}
		}
	}

	var noBoneItem *PhysicsItem

	// ボーンに紐付かない剛体も追加
	for rigidBodyIndex, registered := range registeredRigidBodyIndexes {
		if registered {
			continue
		}
		if rigidBody, err := model.RigidBodies.Get(rigidBodyIndex); err == nil {
			if noBoneItem == nil {
				noBone := pmx.NewBoneByName("No Bone")
				noBoneItem = NewPhysicsItem(model.Bones, noBone, nil, nil)
				tree.AddNode(noBoneItem)
			}
			item := NewPhysicsItem(model.Bones, nil, rigidBody, noBoneItem)
			noBoneItem.AddChild(item)
		}
	}

	// 物理ボーンを持つアイテムのみを保存
	tree.SaveOnlyPhysicsItems()

	return tree
}

func (pm *PhysicsRigidBodyTreeModel) AddNode(node *PhysicsItem) {
	pm.nodes = append(pm.nodes, node)
}

func (pm *PhysicsRigidBodyTreeModel) RootCount() int {
	return len(pm.nodes)
}

func (pm *PhysicsRigidBodyTreeModel) RootAt(index int) walk.TreeItem {
	if index < 0 || index >= len(pm.nodes) {
		return nil
	}
	return pm.nodes[index]
}

func (pm *PhysicsRigidBodyTreeModel) AtByBoneIndex(boneIndex int) walk.TreeItem {
	if boneIndex < 0 {
		return nil
	}

	for _, item := range pm.nodes {
		if found := item.AtByBoneIndex(boneIndex); found != nil {
			return found
		}
	}

	return nil
}

// 物理ボーンを含むツリーだけ残す
func (pm *PhysicsRigidBodyTreeModel) SaveOnlyPhysicsItems() {
	newNodes := make([]*PhysicsItem, 0)
	for _, node := range pm.nodes {
		// 子に物理ボーンがある場合のみ残す
		node.SaveOnlyPhysicsItems()

		if node.HasPhysicsChild() {
			newNodes = append(newNodes, node)
		}
	}
	pm.nodes = newNodes
}

func (pm *PhysicsRigidBodyTreeModel) PublishItemChanged(item walk.TreeItem) {
	if item == nil {
		return
	}

	if _, ok := item.(*PhysicsItem); !ok {
		return
	}

	pm.TreeModelBase.PublishItemChanged(item)

	for _, child := range item.(*PhysicsItem).children {
		pm.PublishItemChanged(child)
	}
}

func (pm *PhysicsRigidBodyTreeModel) Reset() {
	for _, node := range pm.nodes {
		node.Reset()
		pm.PublishItemChanged(node)
	}
}
