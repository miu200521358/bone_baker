package ui

import (
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/walk/pkg/declarative"
)

func NewMenuItems() []declarative.MenuItem {
	return []declarative.MenuItem{
		declarative.Action{
			Text:        "BoneBaker",
			OnTriggered: func() { mlog.ILT("BoneBaker", "%s", mi18n.T("ツール概要説明")) },
		},
		declarative.Separator{},
		declarative.Action{
			Text:        mi18n.T("ワールド物理の調整"),
			OnTriggered: func() { mlog.ILT("ワールド物理の調整", "%s", mi18n.T("ワールド物理の調整説明")) },
		},
		declarative.Action{
			Text:        mi18n.T("モデル物理の調整"),
			OnTriggered: func() { mlog.ILT("モデル物理の調整", "%s", mi18n.T("モデル物理の調整説明")) },
		},
		declarative.Action{
			Text:        mi18n.T("風物理の調整"),
			OnTriggered: func() { mlog.ILT("風物理の調整", "%s", mi18n.T("風物理の調整説明")) },
		},
		declarative.Action{
			Text:        mi18n.T("焼き込み再生"),
			OnTriggered: func() { mlog.ILT("焼き込み再生", "%s", mi18n.T("焼き込み再生説明")) },
		},
		declarative.Action{
			Text: mi18n.T("焼き込み履歴INDEXの選択"),
			OnTriggered: func() {
				mlog.ILT("焼き込み履歴INDEXの選択", "%s", mi18n.T("焼き込み履歴INDEXの選択説明"))
			},
		},
		declarative.Action{
			Text:        mi18n.T("焼き込み結果の出力"),
			OnTriggered: func() { mlog.ILT("焼き込み結果の出力", "%s", mi18n.T("焼き込み結果の出力説明")) },
		},
	}
}
