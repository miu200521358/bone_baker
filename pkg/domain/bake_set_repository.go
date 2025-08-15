package domain

// BakeSetReader BakeSet読み取り専用インターフェース
type BakeSetReader interface {
	// Load ファイルからBakeSetのリストを読み込み
	Load(filePath string) ([]*BakeSet, error)
}

// BakeSetWriter BakeSet書き込み専用インターフェース
type BakeSetWriter interface {
	// Save BakeSetのリストをファイルに保存
	Save(bakeSets []*BakeSet, filePath string) error
}

// BakeSetRepository BakeSetの永続化を担うリポジトリインターフェース
// 読み取りと書き込みの複合インターフェース（下位互換性のため残す）
type BakeSetRepository interface {
	BakeSetReader
	BakeSetWriter
}

// BakeSetPersistenceManager 永続化管理の複合インターフェース
// 必要に応じて読み取り専用や書き込み専用のインターフェースを使用可能
type BakeSetPersistenceManager interface {
	BakeSetRepository
}
