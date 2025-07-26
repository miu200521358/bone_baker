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
			Text:        mi18n.T("物理の調整"),
			OnTriggered: func() { mlog.ILT("物理の調整", "%s", mi18n.T("物理の調整説明")) },
		},
		declarative.Action{
			Text:        mi18n.T("焼き込み結果の確認"),
			OnTriggered: func() { mlog.ILT("焼き込み結果の確認", "%s", mi18n.T("焼き込み結果の確認説明")) },
		},
		declarative.Action{
			Text:        mi18n.T("焼き込み結果の保存"),
			OnTriggered: func() { mlog.ILT("焼き込み結果の保存", "%s", mi18n.T("焼き込み結果の保存説明")) },
		},
	}
}
