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

func (pi *OutputItem) GetCheckedBoneNames() []string {
	var names []string
	if pi.Checked() {
		names = append(names, pi.Text())
	}
	for _, child := range pi.children {
		names = append(names, child.(*OutputItem).GetCheckedBoneNames()...)
	}
	return names
}

type OutputBoneTreeModel struct {
	*walk.TreeModelBase
	nodes []*OutputItem
}

func NewOutputBoneTreeModel(model *pmx.PmxModel) *OutputBoneTreeModel {
	tree := &OutputBoneTreeModel{
		TreeModelBase: &walk.TreeModelBase{},
		nodes:         make([]*OutputItem, 0),
	}

	if model == nil {
		return tree
	}

	// モデルのボーンをツリーに追加
	for _, boneIndex := range model.Bones.LayerSortedIndexes {
		if bone, err := model.Bones.Get(boneIndex); err == nil {
			parent := tree.AtByBoneIndex(bone.ParentIndex)
			item := NewOutputItem(bone, parent)
			if parent == nil {
				tree.AddNode(item)
			} else {
				parent.(*OutputItem).AddChild(item)
			}
		}
	}

	return tree
}

func (pm *OutputBoneTreeModel) GetCheckedBoneNames() []string {
	var names []string
	for _, node := range pm.nodes {
		names = append(names, node.GetCheckedBoneNames()...)
	}
	return names
}

func (pm *OutputBoneTreeModel) AddNode(node *OutputItem) {
	pm.nodes = append(pm.nodes, node)
}

func (pm *OutputBoneTreeModel) RootCount() int {
	return len(pm.nodes)
}

func (pm *OutputBoneTreeModel) RootAt(index int) walk.TreeItem {
	if index < 0 || index >= len(pm.nodes) {
		return nil
	}
	return pm.nodes[index]
}

func (pm *OutputBoneTreeModel) AtByBoneIndex(boneIndex int) walk.TreeItem {
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

func (pm *OutputBoneTreeModel) PublishItemChecked(item walk.TreeItem) {
	if item == nil {
		for _, node := range pm.nodes {
			pm.TreeModelBase.PublishItemChecked(node)
		}
	}

	if _, ok := item.(*OutputItem); !ok {
		return
	}

	pm.TreeModelBase.PublishItemChecked(item)
}

func (pm *OutputBoneTreeModel) SetOutputIkChecked(treeView *walk.TreeView, item walk.TreeItem, checked bool) {
	if item == nil {
		for _, node := range pm.nodes {
			pm.SetOutputIkChecked(treeView, node, checked)
		}
		return
	}

	// 子どもの数を取得
	for i := range item.ChildCount() {
		child := item.ChildAt(i)
		if child == nil {
			continue
		}

		// 出力IKボーンのチェック状態を設定
		if outputItem, ok := child.(*OutputItem); ok {
			if outputItem.AsIk() {
				outputItem.SetChecked(checked)
				treeView.SetChecked(outputItem, checked)
			}
		}

		// 子どもアイテムのチェック状態を設定
		pm.SetOutputIkChecked(treeView, child, checked)
	}
}

func (pm *OutputBoneTreeModel) SetOutputPhysicsChecked(treeView *walk.TreeView, item walk.TreeItem, checked bool) {
	if item == nil {
		for _, node := range pm.nodes {
			pm.SetOutputPhysicsChecked(treeView, node, checked)
		}
		return
	}

	// 子どもの数を取得
	for i := range item.ChildCount() {
		child := item.ChildAt(i)
		if child == nil {
			continue
		}

		// 出力物理ボーンのチェック状態を設定
		if outputItem, ok := child.(*OutputItem); ok {
			if outputItem.AsPhysics() {
				outputItem.SetChecked(checked)
				treeView.SetChecked(outputItem, checked)
			}
		}

		// 子どもアイテムのチェック状態を設定
		pm.SetOutputPhysicsChecked(treeView, child, checked)
	}
}
