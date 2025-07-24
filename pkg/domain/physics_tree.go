package domain

import (
	"fmt"

	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/walk/pkg/walk"
)

type PhysicsItem struct {
	bone           *pmx.Bone
	parent         walk.TreeItem
	children       []walk.TreeItem
	massRatio      float64 // 質量比率
	stiffnessRatio float64 // 硬さ比率
	tensionRatio   float64 // 張り比率
}

func NewPhysicsItem(bone *pmx.Bone, parent walk.TreeItem) *PhysicsItem {
	return &PhysicsItem{
		bone:           bone,
		parent:         parent,
		children:       make([]walk.TreeItem, 0),
		massRatio:      1.0, // 初期値
		stiffnessRatio: 1.0, // 初期値
		tensionRatio:   1.0, // 初期値
	}
}

func (pi *PhysicsItem) Text() string {
	return fmt.Sprintf(mi18n.T("%s (質量: %.2f, 硬さ: %.2f, 張り: %.2f)"),
		pi.bone.Name(), pi.massRatio, pi.stiffnessRatio, pi.tensionRatio)
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
		return pi.bone.HasPhysics()
	}

	hasPhysicsBone := false
	for _, c := range pi.children {
		if c.(*PhysicsItem).HasPhysicsChild() {
			hasPhysicsBone = true
			break
		}
	}

	return hasPhysicsBone || pi.bone.HasPhysics()
}

func (pi *PhysicsItem) Reset() {
	pi.massRatio = 1.0
	pi.stiffnessRatio = 1.0
	pi.tensionRatio = 1.0

	for _, child := range pi.children {
		child.(*PhysicsItem).Reset()
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
	if pi.bone.Index() == boneIndex {
		return pi
	}

	for _, child := range pi.children {
		if found := child.(*PhysicsItem).AtByBoneIndex(boneIndex); found != nil {
			return found
		}
	}

	return nil
}

type PhysicsModel struct {
	*walk.TreeModelBase
	nodes []*PhysicsItem
}

func NewPhysicsModel() *PhysicsModel {
	return &PhysicsModel{
		TreeModelBase: &walk.TreeModelBase{},
		nodes:         make([]*PhysicsItem, 0),
	}
}

func (pm *PhysicsModel) AddNode(node *PhysicsItem) {
	pm.nodes = append(pm.nodes, node)
}

func (pm *PhysicsModel) RootCount() int {
	return len(pm.nodes)
}

func (pm *PhysicsModel) RootAt(index int) walk.TreeItem {
	if index < 0 || index >= len(pm.nodes) {
		return nil
	}
	return pm.nodes[index]
}

func (pm *PhysicsModel) AtByBoneIndex(boneIndex int) walk.TreeItem {
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
func (pm *PhysicsModel) SaveOnlyPhysicsItems() {
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

func (pm *PhysicsModel) PublishItemChanged(item walk.TreeItem) {
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

func (pm *PhysicsModel) Reset() {
	for _, node := range pm.nodes {
		node.Reset()
		pm.PublishItemChanged(node)
	}
}

// GetByID IDでアイテムを取得
func (pm *PhysicsModel) GetByID(id string) walk.TreeItem {
	for _, node := range pm.nodes {
		if found := pm.findByID(node, id); found != nil {
			return found
		}
	}
	return nil
}

// findByID 再帰的にIDでアイテムを検索
func (pm *PhysicsModel) findByID(item *PhysicsItem, id string) walk.TreeItem {
	if item.bone.Name() == id {
		return item
	}

	for _, child := range item.children {
		if found := pm.findByID(child.(*PhysicsItem), id); found != nil {
			return found
		}
	}

	return nil
}

// Children 子要素を取得
func (pi *PhysicsItem) Children() []walk.TreeItem {
	return pi.children
}
