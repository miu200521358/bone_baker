package ui

import (
	"encoding/json"

	"github.com/miu200521358/bone_baker/pkg/domain/entity"
	"github.com/miu200521358/walk/pkg/walk"
)

type OutputTreeItem struct {
	item     *entity.OutputItem // 元のボーンアイテム
	parent   walk.TreeItem      // 親要素(UIあり)
	children []walk.TreeItem    // 子要素(UIあり)
}

func newOutputTreeItem(item *entity.OutputItem, parent walk.TreeItem) *OutputTreeItem {
	treeItem := &OutputTreeItem{
		item:     item,
		parent:   parent,
		children: make([]walk.TreeItem, 0),
	}

	for _, child := range item.Children {
		childItem := newOutputTreeItem(child, treeItem)
		treeItem.AddChild(childItem)
	}

	return treeItem
}

func (pi *OutputTreeItem) Text() string {
	if pi.item.Bone != nil {
		return pi.item.Bone.Name()
	}

	return "Unknown"
}

func (pi *OutputTreeItem) Parent() walk.TreeItem {
	if pi.parent == nil {
		return nil
	}
	return pi.parent
}

func (pi *OutputTreeItem) AddChild(child walk.TreeItem) {
	pi.children = append(pi.children, child)
}

func (pi *OutputTreeItem) Reset() {
	pi.item.Checked = false

	for _, child := range pi.children {
		child.(*OutputTreeItem).Reset()
	}
}

func (pi *OutputTreeItem) ChildCount() int {
	return len(pi.children)
}

func (pi *OutputTreeItem) HasChild() bool {
	return len(pi.children) > 0
}

func (pi *OutputTreeItem) ChildAt(index int) walk.TreeItem {
	if index < 0 || index >= len(pi.children) {
		return nil
	}
	return pi.children[index]
}

func (pi *OutputTreeItem) AtByBoneIndex(boneIndex int) *OutputTreeItem {
	if pi.item.Bone == nil {
		return nil
	}

	if pi.item.Bone != nil && pi.item.Bone.Index() == boneIndex {
		return pi
	}

	for _, child := range pi.children {
		if found := child.(*OutputTreeItem).AtByBoneIndex(boneIndex); found != nil {
			return found
		}
	}

	return nil
}

type OutputTreeModel struct {
	*walk.TreeModelBase
	Record *entity.PhysicsRecord // 出力レコード
	Nodes  []*OutputTreeItem     // 出力ノード
}

func newOutputTreeModel(rigidBodyRecord *entity.OutputRecord) *OutputTreeModel {
	tree := &OutputTreeModel{
		TreeModelBase: &walk.TreeModelBase{},
		Nodes:         make([]*OutputTreeItem, 0),
	}

	for _, item := range rigidBodyRecord.Tree.Items {
		treeItem := newOutputTreeItem(item, nil)
		tree.AddNode(treeItem)
	}

	return tree
}

func (pm *OutputTreeModel) MarshalJSON() ([]byte, error) {
	// 変更のあったノードのみを収集
	modifiedNodes := pm.CheckedNodes(nil)

	// 変更のあったフィールドのみを含める
	return json.Marshal(&struct {
		Nodes []*OutputTreeItem `json:"Nodes"`
	}{
		Nodes: modifiedNodes,
	})
}

func (pm *OutputTreeModel) CheckedNodes(node *OutputTreeItem) []*OutputTreeItem {
	checkedNodes := make([]*OutputTreeItem, 0)
	if node == nil {
		for _, node := range pm.Nodes {
			checkedNodes = append(checkedNodes, pm.CheckedNodes(node)...)
		}
		return checkedNodes
	}

	if node.item.Checked {
		checkedNodes = append(checkedNodes, node)
	}

	for _, child := range node.children {
		checkedNodes = append(checkedNodes, pm.CheckedNodes(child.(*OutputTreeItem))...)
	}

	return checkedNodes
}

func (pm *OutputTreeModel) UpdateCheckedNodes(node *OutputTreeItem, modifiedNodes []*OutputTreeItem) {
	if node == nil {
		for _, n := range pm.Nodes {
			pm.UpdateCheckedNodes(n, modifiedNodes)
		}
		return
	}

	for _, mNodes := range modifiedNodes {
		// 同じボーンインデックスと名前を持つノードの場合、更新
		if mNodes.item.Bone.Index() == node.item.Bone.Index() && mNodes.item.Bone.Name() == node.item.Bone.Name() {
			node.item.Checked = mNodes.item.Checked
		}
	}

	for _, child := range node.children {
		pm.UpdateCheckedNodes(child.(*OutputTreeItem), modifiedNodes)
	}
}

func (pm *OutputTreeModel) AddNode(node *OutputTreeItem) {
	pm.Nodes = append(pm.Nodes, node)
}

func (pm *OutputTreeModel) RootCount() int {
	return len(pm.Nodes)
}

func (pm *OutputTreeModel) RootAt(index int) walk.TreeItem {
	if index < 0 || index >= len(pm.Nodes) {
		return nil
	}
	return pm.Nodes[index]
}

func (pm *OutputTreeModel) AtByBoneIndex(boneIndex int) walk.TreeItem {
	if boneIndex < 0 {
		return nil
	}

	for _, item := range pm.Nodes {
		if found := item.AtByBoneIndex(boneIndex); found != nil {
			return found
		}
	}

	return nil
}

func (pm *OutputTreeModel) PublishItemChanged(item walk.TreeItem) {
	if item == nil {
		return
	}

	if _, ok := item.(*OutputTreeItem); !ok {
		return
	}

	pm.TreeModelBase.PublishItemChanged(item)

	for _, child := range item.(*OutputTreeItem).children {
		pm.PublishItemChanged(child)
	}
}

func (pm *OutputTreeModel) Reset() {
	for _, node := range pm.Nodes {
		node.Reset()
		pm.PublishItemChanged(node)
	}
}

func (pm *OutputTreeModel) SetOutputIkChecked(treeView *walk.TreeView, item walk.TreeItem, checked bool) {
	if item == nil {
		for _, node := range pm.Nodes {
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
		if outputItem, ok := child.(*OutputTreeItem); ok {
			if outputItem.item.AsIk() {
				outputItem.item.Checked = checked
				treeView.SetChecked(outputItem, checked)
			}
		}

		// 子どもアイテムのチェック状態を設定
		pm.SetOutputIkChecked(treeView, child, checked)
	}
}

func (pm *OutputTreeModel) SetOutputStandardChecked(treeView *walk.TreeView, item walk.TreeItem, checked bool) {
	if item == nil {
		for _, node := range pm.Nodes {
			pm.SetOutputStandardChecked(treeView, node, checked)
		}
		return
	}

	// 子どもの数を取得
	for i := range item.ChildCount() {
		child := item.ChildAt(i)
		if child == nil {
			continue
		}

		// 出力準標準ボーンのチェック状態を設定
		if outputItem, ok := child.(*OutputTreeItem); ok {
			if outputItem.item.AsStandard() {
				outputItem.item.Checked = checked
				treeView.SetChecked(outputItem, checked)
			}
		}

		// 子どもアイテムのチェック状態を設定
		pm.SetOutputStandardChecked(treeView, child, checked)
	}
}

func (pm *OutputTreeModel) SetOutputFingerChecked(treeView *walk.TreeView, item walk.TreeItem, checked bool) {
	if item == nil {
		for _, node := range pm.Nodes {
			pm.SetOutputFingerChecked(treeView, node, checked)
		}
		return
	}

	// 子どもの数を取得
	for i := range item.ChildCount() {
		child := item.ChildAt(i)
		if child == nil {
			continue
		}

		// 出力指ボーンのチェック状態を設定
		if outputItem, ok := child.(*OutputTreeItem); ok {
			if outputItem.item.AsFinger() {
				outputItem.item.Checked = checked
				treeView.SetChecked(outputItem, checked)
			}
		}

		// 子どもアイテムのチェック状態を設定
		pm.SetOutputFingerChecked(treeView, child, checked)
	}
}

func (pm *OutputTreeModel) SetOutputPhysicsChecked(treeView *walk.TreeView, item walk.TreeItem, checked bool) {
	if item == nil {
		for _, node := range pm.Nodes {
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
		if outputItem, ok := child.(*OutputTreeItem); ok {
			if outputItem.item.AsDynamicPhysics() {
				outputItem.item.Checked = checked
				treeView.SetChecked(outputItem, checked)
			}
		}

		// 子どもアイテムのチェック状態を設定
		pm.SetOutputPhysicsChecked(treeView, child, checked)
	}
}
