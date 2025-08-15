package domain

// BakeSetRepository BakeSetの永続化を担うリポジトリインターフェース
// ドメイン層で定義し、インフラストラクチャ層で実装する
type BakeSetRepository interface {
	// Save BakeSetのリストをファイルに保存
	Save(bakeSets []*BakeSet, filePath string) error

	// Load ファイルからBakeSetのリストを読み込み
	Load(filePath string) ([]*BakeSet, error)
}
