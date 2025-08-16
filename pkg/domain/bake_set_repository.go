package domain

// BakeSetRepository BakeSetの永続化を担うリポジトリインターフェース
type BakeSetRepository interface {
	// Save BakeSetのリストをファイルに保存
	Save(bakeSets []*BakeSet, filePath string) error
	// Load ファイルからBakeSetのリストを読み込み
	Load(filePath string) ([]*BakeSet, error)
}
