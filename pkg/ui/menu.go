package ui

import (
	"github.com/miu200521358/mlib_go/pkg/config/mi18n"
	"github.com/miu200521358/mlib_go/pkg/config/mlog"
	"github.com/miu200521358/walk/pkg/declarative"
)

func NewMenuItems() []declarative.MenuItem {
	return []declarative.MenuItem{
		declarative.Action{
			Text:        mi18n.T("概要"),
			OnTriggered: func() { mlog.ILT(mi18n.T("概要"), "%s", mi18n.T("概要説明")) },
		},
		declarative.Action{
			Text:        mi18n.T("名称置換"),
			OnTriggered: func() { mlog.ILT(mi18n.T("名称置換"), "%s", mi18n.T("名称置換使い方")) },
		},
		declarative.Action{
			Text:        mi18n.T("Csv出力"),
			OnTriggered: func() { mlog.ILT(mi18n.T("Csv出力"), "%s", mi18n.T("Csv出力使い方")) },
		},
		declarative.Action{
			Text:        mi18n.T("Csv追加"),
			OnTriggered: func() { mlog.ILT(mi18n.T("Csv追加"), "%s", mi18n.T("Csv追加使い方")) },
		},
	}
}
