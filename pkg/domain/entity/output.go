package entity

import (
	"slices"
	"strings"

	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/domain/pmx"
)

type OutputBoneFlag = int

const (
	OutputBoneFlagEmpty    OutputBoneFlag = 0 // 出力無し
	OutputBoneFlagOriginal OutputBoneFlag = 1 // 元モーション登録
	OutputBoneFlagBake     OutputBoneFlag = 2 // 出力モーション登録
	OutputBoneFlagReduce   OutputBoneFlag = 4 // 間引き出力
)

type OutputRecord struct {
	StartFrame float32     `json:"start_frame"` // 区間開始フレーム
	EndFrame   float32     `json:"end_frame"`   // 区間終了フレーム
	Reduce     bool        `json:"reduce"`      // 間引き有無
	Tree       *OutputTree `json:"items"`       // ボーンアイテム一覧
}

func NewOutputRecord(startFrame, endFrame float32, model *pmx.PmxModel) *OutputRecord {
	return &OutputRecord{
		StartFrame: startFrame,
		EndFrame:   endFrame,
		Tree:       newOutputTree(model),
	}
}

func (r *OutputRecord) ItemNames() string {
	boneNames := r.ItemBoneNames()

	if len(boneNames) == 0 {
		return mi18n.T("出力対象ボーンなし")
	}

	if len(boneNames) <= 6 {
		return strings.Join(boneNames, ", ")
	}

	return strings.Join(r.ItemBoneNames()[:6], ", ") + "..."
}

func (r *OutputRecord) ItemBoneNames() []string {
	var names []string
	for _, item := range r.Tree.Items {
		names = append(names, item.ItemBoneNames()...)
	}

	// 重複削除
	names = slices.Compact(names)

	return names
}

type OutputTree struct {
	Items []*OutputItem
}

func newOutputTree(model *pmx.PmxModel) *OutputTree {
	items := &OutputTree{}

	for _, boneIndex := range model.Bones.LayerSortedIndexes {
		if bone, err := model.Bones.Get(boneIndex); err == nil {
			parent := items.AtByBoneIndex(bone.ParentIndex)
			item := newOutputItem(bone, parent)
			if parent == nil {
				items.AddNode(item)
			} else {
				parent.AddChild(item)
			}
		}
	}

	return items
}

func (oi *OutputItem) AsIk() bool {
	return oi.Bone.IsVisible() && (oi.Bone.IsIK() || len(oi.Bone.IkLinkBoneIndexes) > 0)
}

// 全親・足D・指・目以外の準標準ボーン
func (oi *OutputItem) AsStandard() bool {
	return oi.Bone.IsVisible() && oi.Bone.Config() != nil && oi.Bone.Config().IsStandard && !oi.Bone.Config().IsFinger() && !oi.Bone.Config().IsLegD() && !oi.Bone.Config().IsRoot() && !oi.Bone.Config().IsEye()
}

func (oi *OutputItem) AsFinger() bool {
	return oi.Bone.IsVisible() && oi.Bone.Config() != nil && oi.Bone.Config().IsStandard && oi.Bone.Config().IsFinger()
}

func (oi *OutputItem) AsDynamicPhysics() bool {
	return oi.Bone.IsVisible() && oi.Bone.HasDynamicPhysics()
}

func (r *OutputTree) AddNode(item *OutputItem) {
	r.Items = append(r.Items, item)
}

func (r OutputTree) AtByBoneIndex(boneIndex int) *OutputItem {
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

type OutputItem struct {
	Bone     *pmx.Bone     // ボーン情報
	Checked  bool          // チェック有無
	Parent   *OutputItem   `json:"parent"`   // 親ボーンアイテム
	Children []*OutputItem `json:"Children"` // 子ボーンアイテム
}

func newOutputItem(bone *pmx.Bone, parent *OutputItem) *OutputItem {
	item := &OutputItem{
		Bone:     bone,
		Checked:  false,
		Parent:   parent,
		Children: []*OutputItem{},
	}

	return item
}

func (oi *OutputItem) ItemBoneNames() []string {
	names := make([]string, 0)
	if oi.Bone != nil && oi.Bone.DisplaySlotIndex >= 0 && oi.Checked {
		// 表示枠に登録されているボーンのみ対象とする
		names = append(names, oi.Bone.Name())
	}

	for _, child := range oi.Children {
		names = append(names, child.ItemBoneNames()...)
	}

	return names
}

func (oi *OutputItem) AtByBoneIndex(boneIndex int) *OutputItem {
	if oi.Bone == nil {
		return nil
	}

	if oi.Bone.Index() == boneIndex {
		return oi
	}

	for _, child := range oi.Children {
		if found := child.AtByBoneIndex(boneIndex); found != nil {
			return found
		}
	}

	return nil
}

func (oi *OutputItem) AddChild(child *OutputItem) {
	oi.Children = append(oi.Children, child)
}

func (oi *OutputItem) ChildCount() int {
	return len(oi.Children)
}

func (oi *OutputItem) HasChild() bool {
	return len(oi.Children) > 0
}

func (oi *OutputItem) ChildAt(index int) *OutputItem {
	if index < 0 || index >= len(oi.Children) {
		return nil
	}
	return oi.Children[index]
}
