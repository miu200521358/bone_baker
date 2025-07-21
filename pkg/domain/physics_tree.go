package domain

import (
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/walk/pkg/walk"
)

type PhysicsItem struct {
	bone     *pmx.Bone
	parent   walk.TreeItem
	children []walk.TreeItem
}

func NewPhysicsItem(bone *pmx.Bone, parent walk.TreeItem) *PhysicsItem {

	return &PhysicsItem{
		bone:     bone,
		parent:   parent,
		children: make([]walk.TreeItem, 0),
	}
}

func (pi *PhysicsItem) Text() string {
	return pi.bone.Name()
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
