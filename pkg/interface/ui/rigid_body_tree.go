package ui

import (
	"encoding/json"
	"fmt"

	"github.com/miu200521358/bone_baker/pkg/domain/entity"
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/domain/mmath"
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
	"github.com/miu200521358/walk/pkg/walk"
)

type RigidBodyTreeItem struct {
	item     *entity.RigidBodyItem // 元の剛体アイテム
	parent   walk.TreeItem         // 親要素(UIあり)
	children []walk.TreeItem       // 子要素(UIあり)
}

func newRigidBodyTreeItem(item *entity.RigidBodyItem, parent walk.TreeItem) *RigidBodyTreeItem {
	treeItem := &RigidBodyTreeItem{
		item:     item,
		parent:   parent,
		children: make([]walk.TreeItem, 0),
	}

	for _, child := range item.Children {
		childItem := newRigidBodyTreeItem(child, treeItem)
		treeItem.AddChild(childItem)
	}

	return treeItem
}

func (pi *RigidBodyTreeItem) Text() string {
	if pi.item.RigidBody == nil {
		return fmt.Sprintf(mi18n.T("%s (剛体なし)"), pi.item.Bone.Name())
	}

	var nameText string
	if pi.item.RigidBody != nil {
		nameText = pi.item.RigidBody.Name()
	} else if pi.item.Bone != nil {
		nameText = pi.item.Bone.Name()
	} else {
		nameText = "Unknown"
	}

	var sizeText string
	switch pi.item.RigidBody.ShapeType {
	case pmx.SHAPE_SPHERE:
		sizeText = fmt.Sprintf(mi18n.T("半径: %.2f"), pi.item.SizeRatio.X)
	case pmx.SHAPE_BOX:
		sizeText = fmt.Sprintf(mi18n.T("幅: %.2f, 高さ: %.2f, 奥行: %.2f"), pi.item.SizeRatio.X, pi.item.SizeRatio.Y, pi.item.SizeRatio.Z)
	case pmx.SHAPE_CAPSULE:
		sizeText = fmt.Sprintf(mi18n.T("半径: %.2f, 高さ: %.2f"), pi.item.SizeRatio.X, pi.item.SizeRatio.Y)
	}

	return fmt.Sprintf(mi18n.T("%s (大きさ倍率: [%s], 質量: %.2f, 硬さ: %.2f, 張り: %.2f)"),
		nameText, sizeText, pi.item.MassRatio, pi.item.StiffnessRatio, pi.item.TensionRatio)
}

func (pi *RigidBodyTreeItem) Parent() walk.TreeItem {
	if pi.parent == nil {
		return nil
	}
	return pi.parent
}

func (pi *RigidBodyTreeItem) AddChild(child walk.TreeItem) {
	pi.children = append(pi.children, child)
}

func (pi *RigidBodyTreeItem) HasPhysicsChild() bool {
	if len(pi.children) == 0 {
		return pi.item.RigidBody != nil
	}

	hasPhysicsRigidBody := false
	for _, c := range pi.children {
		if c.(*RigidBodyTreeItem).HasPhysicsChild() {
			hasPhysicsRigidBody = true
			break
		}
	}

	return hasPhysicsRigidBody || pi.item.RigidBody != nil
}

func (pi *RigidBodyTreeItem) Reset() {
	pi.item.SizeRatio = &mmath.MVec3{X: 1.0, Y: 1.0, Z: 1.0} // 大きさ比率を初期化
	pi.item.MassRatio = 1.0
	pi.item.StiffnessRatio = 1.0
	pi.item.TensionRatio = 1.0

	for _, child := range pi.children {
		child.(*RigidBodyTreeItem).Reset()
	}
}

func (pi *RigidBodyTreeItem) CalcSizeX(x float64) {
	pi.item.SizeRatio.X = x
	pi.item.Modified = true

	for _, child := range pi.children {
		child.(*RigidBodyTreeItem).CalcSizeX(x)
	}
}

func (pi *RigidBodyTreeItem) CalcSizeY(y float64) {
	pi.item.SizeRatio.Y = y
	pi.item.Modified = true

	for _, child := range pi.children {
		child.(*RigidBodyTreeItem).CalcSizeY(y)
	}
}

func (pi *RigidBodyTreeItem) CalcSizeZ(z float64) {
	pi.item.SizeRatio.Z = z
	pi.item.Modified = true

	for _, child := range pi.children {
		child.(*RigidBodyTreeItem).CalcSizeZ(z)
	}
}

func (pi *RigidBodyTreeItem) CalcPositionX(x float64) {
	pi.item.Position.X = x
	pi.item.Modified = true

	// for _, child := range pi.children {
	// 	child.(*RigidBodyTreeItem).CalcPositionX(x)
	// }
}

func (pi *RigidBodyTreeItem) CalcPositionY(y float64) {
	pi.item.Position.Y = y
	pi.item.Modified = true

	// for _, child := range pi.children {
	// 	child.(*RigidBodyTreeItem).CalcPositionY(y)
	// }
}

func (pi *RigidBodyTreeItem) CalcPositionZ(z float64) {
	pi.item.Position.Z = z
	pi.item.Modified = true

	// for _, child := range pi.children {
	// 	child.(*RigidBodyTreeItem).CalcPositionZ(z)
	// }
}

func (pi *RigidBodyTreeItem) CalcMass(massRatio float64) {
	pi.item.MassRatio = massRatio
	pi.item.Modified = true

	for _, child := range pi.children {
		child.(*RigidBodyTreeItem).CalcMass(massRatio)
	}
}

func (pi *RigidBodyTreeItem) CalcStiffness(stiffnessRatio float64) {
	pi.item.StiffnessRatio = stiffnessRatio
	pi.item.Modified = true

	for _, child := range pi.children {
		child.(*RigidBodyTreeItem).CalcStiffness(stiffnessRatio)
	}
}

func (pi *RigidBodyTreeItem) CalcTension(tensionRatio float64) {
	pi.item.TensionRatio = tensionRatio
	pi.item.Modified = true

	for _, child := range pi.children {
		child.(*RigidBodyTreeItem).CalcTension(tensionRatio)
	}
}

func (pi *RigidBodyTreeItem) SaveOnlyPhysicsItems() {
	newChildren := make([]walk.TreeItem, 0)
	for _, child := range pi.children {
		child.(*RigidBodyTreeItem).SaveOnlyPhysicsItems()

		if child.(*RigidBodyTreeItem).HasPhysicsChild() {
			newChildren = append(newChildren, child)
		}
	}

	pi.children = newChildren
}

func (pi *RigidBodyTreeItem) ChildCount() int {
	return len(pi.children)
}

func (pi *RigidBodyTreeItem) HasChild() bool {
	return len(pi.children) > 0
}

func (pi *RigidBodyTreeItem) ChildAt(index int) walk.TreeItem {
	if index < 0 || index >= len(pi.children) {
		return nil
	}
	return pi.children[index]
}

func (pi *RigidBodyTreeItem) AtByBoneIndex(boneIndex int) *RigidBodyTreeItem {
	if pi.item.Bone == nil {
		return nil
	}

	if pi.item.Bone != nil && pi.item.Bone.Index() == boneIndex {
		return pi
	}

	for _, child := range pi.children {
		if found := child.(*RigidBodyTreeItem).AtByBoneIndex(boneIndex); found != nil {
			return found
		}
	}

	return nil
}

func (pi *RigidBodyTreeItem) AtByRigidBodyIndex(rigidBodyIndex int) *RigidBodyTreeItem {
	if pi.item.RigidBody == nil {
		for _, child := range pi.children {
			if found := child.(*RigidBodyTreeItem).AtByRigidBodyIndex(rigidBodyIndex); found != nil {
				return found
			}
		}

		return nil
	}

	if pi.item.RigidBody.Index() == rigidBodyIndex {
		return pi
	}

	for _, child := range pi.children {
		if found := child.(*RigidBodyTreeItem).AtByRigidBodyIndex(rigidBodyIndex); found != nil {
			return found
		}
	}

	return nil
}

type RigidBodyTreeModel struct {
	*walk.TreeModelBase
	Record *entity.PhysicsRecord // 物理レコード
	Nodes  []*RigidBodyTreeItem  // 物理剛体ノード
}

func newRigidBodyTreeModel(rigidBodyRecord *entity.RigidBodyRecord) *RigidBodyTreeModel {
	tree := &RigidBodyTreeModel{
		TreeModelBase: &walk.TreeModelBase{},
		Nodes:         make([]*RigidBodyTreeItem, 0),
	}

	for _, item := range rigidBodyRecord.Tree.Items {
		treeItem := newRigidBodyTreeItem(item, nil)
		tree.AddNode(treeItem)
	}

	return tree
}

func (pm *RigidBodyTreeModel) MarshalJSON() ([]byte, error) {
	// 変更のあったノードのみを収集
	modifiedNodes := pm.ModifiedNodes(nil)

	// 変更のあったフィールドのみを含める
	return json.Marshal(&struct {
		Nodes []*RigidBodyTreeItem `json:"nodes"`
	}{
		Nodes: modifiedNodes,
	})
}

func (pm *RigidBodyTreeModel) ModifiedNodes(node *RigidBodyTreeItem) []*RigidBodyTreeItem {
	modifiedNodes := make([]*RigidBodyTreeItem, 0)
	if node == nil {
		for _, node := range pm.Nodes {
			modifiedNodes = append(modifiedNodes, pm.ModifiedNodes(node)...)
		}
		return modifiedNodes
	}

	if node.item.Modified {
		modifiedNodes = append(modifiedNodes, node)
	}

	for _, child := range node.children {
		modifiedNodes = append(modifiedNodes, pm.ModifiedNodes(child.(*RigidBodyTreeItem))...)
	}

	return modifiedNodes
}

func (pm *RigidBodyTreeModel) UpdateModifiedNodes(node *RigidBodyTreeItem, modifiedNodes []*RigidBodyTreeItem) {
	if node == nil {
		for _, n := range pm.Nodes {
			pm.UpdateModifiedNodes(n, modifiedNodes)
		}
		return
	}

	for _, mNodes := range modifiedNodes {
		// 同じ剛体インデックスと名前を持つノードの場合、更新
		if mNodes.item.RigidBody.Index() == node.item.RigidBody.Index() && mNodes.item.RigidBody.Name() == node.item.RigidBody.Name() {
			node.item.SizeRatio = mNodes.item.SizeRatio
			node.item.MassRatio = mNodes.item.MassRatio
			node.item.StiffnessRatio = mNodes.item.StiffnessRatio
			node.item.TensionRatio = mNodes.item.TensionRatio
			node.item.Modified = mNodes.item.Modified
		}
	}

	for _, child := range node.children {
		pm.UpdateModifiedNodes(child.(*RigidBodyTreeItem), modifiedNodes)
	}
}

func (pm *RigidBodyTreeModel) AddNode(node *RigidBodyTreeItem) {
	pm.Nodes = append(pm.Nodes, node)
}

func (pm *RigidBodyTreeModel) RootCount() int {
	return len(pm.Nodes)
}

func (pm *RigidBodyTreeModel) RootAt(index int) walk.TreeItem {
	if index < 0 || index >= len(pm.Nodes) {
		return nil
	}
	return pm.Nodes[index]
}

func (pm *RigidBodyTreeModel) AtByBoneIndex(boneIndex int) walk.TreeItem {
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

func (pm *RigidBodyTreeModel) AtByRigidBodyIndex(rigidBodyIndex int) walk.TreeItem {
	if rigidBodyIndex < 0 {
		return nil
	}

	for _, item := range pm.Nodes {
		if found := item.AtByRigidBodyIndex(rigidBodyIndex); found != nil {
			return found
		}
	}

	return nil
}

// 物理ボーンを含むツリーだけ残す
func (pm *RigidBodyTreeModel) SaveOnlyPhysicsItems() {
	newNodes := make([]*RigidBodyTreeItem, 0)
	for _, node := range pm.Nodes {
		// 子に物理ボーンがある場合のみ残す
		node.SaveOnlyPhysicsItems()

		if node.HasPhysicsChild() {
			newNodes = append(newNodes, node)
		}
	}
	pm.Nodes = newNodes
}

func (pm *RigidBodyTreeModel) PublishItemChanged(item walk.TreeItem) {
	if item == nil {
		return
	}

	if _, ok := item.(*RigidBodyTreeItem); !ok {
		return
	}

	pm.TreeModelBase.PublishItemChanged(item)

	for _, child := range item.(*RigidBodyTreeItem).children {
		pm.PublishItemChanged(child)
	}
}

func (pm *RigidBodyTreeModel) Reset() {
	for _, node := range pm.Nodes {
		node.Reset()
		pm.PublishItemChanged(node)
	}
}
