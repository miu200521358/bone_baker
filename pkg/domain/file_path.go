package domain

import (
	"fmt"

	"github.com/miu200521358/mlib_go/pkg/infrastructure/mfile"
)

// FilePath ファイルパスのValue Object
type FilePath struct {
	value string
}

// NewFilePath ファイルパスのコンストラクタ
func NewFilePath(path string) *FilePath {
	return &FilePath{value: path}
}

// Value パス文字列を取得
func (fp *FilePath) Value() string {
	return fp.value
}

// IsEmpty パスが空かどうか
func (fp *FilePath) IsEmpty() bool {
	return fp.value == ""
}

// String 文字列表現
func (fp *FilePath) String() string {
	return fp.value
}

// CreateOutputPath 出力パスを生成
func (fp *FilePath) CreateOutputPath(suffix string) *FilePath {
	if fp.IsEmpty() {
		return NewFilePath("")
	}

	outputPath := mfile.CreateOutputPath(fp.value, suffix)
	return NewFilePath(outputPath)
}

// CreateOutputPathWithFileName ファイル名を含む出力パスを生成
func (fp *FilePath) CreateOutputPathWithFileName(basePath string, fileName string) *FilePath {
	if fp.IsEmpty() {
		return NewFilePath("")
	}

	outputPath := mfile.CreateOutputPath(basePath, fmt.Sprintf("BB_%s", fileName))
	return NewFilePath(outputPath)
}

// SplitPath パスを分割
func (fp *FilePath) SplitPath() (dirPath, fileName, ext string) {
	if fp.IsEmpty() {
		return "", "", ""
	}
	return mfile.SplitPath(fp.value)
}
