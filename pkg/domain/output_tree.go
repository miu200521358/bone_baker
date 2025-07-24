package domain

import (
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/walk/pkg/walk"
)

type OutputItem struct {
	bone     *pmx.Bone
	parent   walk.TreeItem
	children []walk.TreeItem
	checked  bool
}

func NewOutputItem(bone *pmx.Bone, parent walk.TreeItem) *OutputItem {
	return &OutputItem{
		bone:     bone,
		parent:   parent,
		children: make([]walk.TreeItem, 0),
	}
}

func (pi *OutputItem) AsIk() bool {
	return pi.bone.IsIK() || len(pi.bone.IkLinkBoneIndexes) > 0
}

func (pi *OutputItem) AsPhysics() bool {
	return pi.bone.HasPhysics()
}

func (pi *OutputItem) SetChecked(checked bool) {
	pi.checked = checked
}

func (pi *OutputItem) Checked() bool {
	return pi.checked
}

func (pi *OutputItem) Text() string {
	return pi.bone.Name()
}

func (pi *OutputItem) Parent() walk.TreeItem {
	if pi.parent == nil {
		return nil
	}
	return pi.parent
}

func (pi *OutputItem) AddChild(child walk.TreeItem) {
	pi.children = append(pi.children, child)
}

func (pi *OutputItem) ChildCount() int {
	return len(pi.children)
}

func (pi *OutputItem) HasChild() bool {
	return len(pi.children) > 0
}

func (pi *OutputItem) ChildAt(index int) walk.TreeItem {
	if index < 0 || index >= len(pi.children) {
		return nil
	}
	return pi.children[index]
}

func (pi *OutputItem) AtByBoneIndex(boneIndex int) *OutputItem {
	if pi.bone.Index() == boneIndex {
		return pi
	}

	for _, child := range pi.children {
		if found := child.(*OutputItem).AtByBoneIndex(boneIndex); found != nil {
			return found
		}
	}

	return nil
}

type OutputModel struct {
	*walk.TreeModelBase
	nodes []*OutputItem
}

func NewOutputModel() *OutputModel {
	return &OutputModel{
		TreeModelBase: &walk.TreeModelBase{},
		nodes:         make([]*OutputItem, 0),
	}
}

func (pm *OutputModel) AddNode(node *OutputItem) {
	pm.nodes = append(pm.nodes, node)
}

func (pm *OutputModel) RootCount() int {
	return len(pm.nodes)
}

func (pm *OutputModel) RootAt(index int) walk.TreeItem {
	if index < 0 || index >= len(pm.nodes) {
		return nil
	}
	return pm.nodes[index]
}

func (pm *OutputModel) AtByBoneIndex(boneIndex int) walk.TreeItem {
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

func (pm *OutputModel) PublishItemChecked(item walk.TreeItem) {
	if item == nil {
		return
	}

	if _, ok := item.(*OutputItem); !ok {
		return
	}

	pm.TreeModelBase.PublishItemChecked(item)
}

// GetByID IDでアイテムを取得
func (pm *OutputModel) GetByID(id string) walk.TreeItem {
	for _, node := range pm.nodes {
		if found := pm.findByID(node, id); found != nil {
			return found
		}
	}
	return nil
}

// findByID 再帰的にIDでアイテムを検索
func (pm *OutputModel) findByID(item *OutputItem, id string) walk.TreeItem {
	if item.bone.Name() == id {
		return item
	}

	for _, child := range item.children {
		if found := pm.findByID(child.(*OutputItem), id); found != nil {
			return found
		}
	}

	return nil
}

// Children 子要素を取得
func (oi *OutputItem) Children() []walk.TreeItem {
	return oi.children
}

// GetAllNodes すべてのノードを取得
func (om *OutputModel) GetAllNodes() []*OutputItem {
	return om.nodes
}
