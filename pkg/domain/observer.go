package domain

// Observer 観察者パターンのObserverインターフェース
type Observer interface {
	Update(event *BakeEvent)
	GetObserverID() string
}

// BakeEvent 焼き込み処理のイベント
type BakeEvent struct {
	EventType BakeEventType
	Message   string
	Progress  float64 // 0.0 - 1.0
	Data      interface{}
	Timestamp int64
}

// BakeEventType 焼き込みイベントの種類
type BakeEventType string

const (
	BakeEventTypeStarted   BakeEventType = "started"
	BakeEventTypeProgress  BakeEventType = "progress"
	BakeEventTypeCompleted BakeEventType = "completed"
	BakeEventTypeError     BakeEventType = "error"
	BakeEventTypeCanceled  BakeEventType = "canceled"
)

// Subject 観察者パターンのSubjectインターフェース
type Subject interface {
	Attach(observer Observer)
	Detach(observer Observer)
	Notify(event *BakeEvent)
}

// BakeNotifier 焼き込み処理の通知管理
type BakeNotifier struct {
	observers []Observer
}

// NewBakeNotifier 通知管理のコンストラクタ
func NewBakeNotifier() *BakeNotifier {
	return &BakeNotifier{
		observers: make([]Observer, 0),
	}
}

// Attach 観察者を追加
func (n *BakeNotifier) Attach(observer Observer) {
	n.observers = append(n.observers, observer)
}

// Detach 観察者を削除
func (n *BakeNotifier) Detach(observer Observer) {
	for i, obs := range n.observers {
		if obs.GetObserverID() == observer.GetObserverID() {
			n.observers = append(n.observers[:i], n.observers[i+1:]...)
			break
		}
	}
}

// Notify すべての観察者に通知
func (n *BakeNotifier) Notify(event *BakeEvent) {
	for _, observer := range n.observers {
		observer.Update(event)
	}
}

// ProgressObserver 進捗表示用の観察者
type ProgressObserver struct {
	id              string
	progressHandler func(*BakeEvent)
}

// NewProgressObserver 進捗観察者のコンストラクタ
func NewProgressObserver(id string, handler func(*BakeEvent)) *ProgressObserver {
	return &ProgressObserver{
		id:              id,
		progressHandler: handler,
	}
}

// Update イベント更新処理
func (o *ProgressObserver) Update(event *BakeEvent) {
	if o.progressHandler != nil {
		o.progressHandler(event)
	}
}

// GetObserverID 観察者IDを取得
func (o *ProgressObserver) GetObserverID() string {
	return o.id
}

// LogObserver ログ出力用の観察者
type LogObserver struct {
	id      string
	logFunc func(message string)
}

// NewLogObserver ログ観察者のコンストラクタ
func NewLogObserver(id string, logFunc func(string)) *LogObserver {
	return &LogObserver{
		id:      id,
		logFunc: logFunc,
	}
}

// Update イベント更新処理
func (o *LogObserver) Update(event *BakeEvent) {
	if o.logFunc != nil {
		o.logFunc(event.Message)
	}
}

// GetObserverID 観察者IDを取得
func (o *LogObserver) GetObserverID() string {
	return o.id
}
