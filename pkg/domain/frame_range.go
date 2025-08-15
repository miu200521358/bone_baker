package domain

import (
	"errors"
	"fmt"
)

// FrameRange フレーム範囲のValue Object
type FrameRange struct {
	startFrame float32
	endFrame   float32
}

// NewFrameRange フレーム範囲のコンストラクタ
func NewFrameRange(startFrame, endFrame float32) (*FrameRange, error) {
	if startFrame < 0 {
		return nil, errors.New("開始フレームは0以上である必要があります")
	}
	if endFrame < startFrame {
		return nil, fmt.Errorf("終了フレーム(%f)は開始フレーム(%f)以上である必要があります", endFrame, startFrame)
	}

	return &FrameRange{
		startFrame: startFrame,
		endFrame:   endFrame,
	}, nil
}

// NewFrameRangeUnsafe バリデーションなしでフレーム範囲を作成（既存コードとの互換性用）
func NewFrameRangeUnsafe(startFrame, endFrame float32) *FrameRange {
	return &FrameRange{
		startFrame: startFrame,
		endFrame:   endFrame,
	}
}

// StartFrame 開始フレームを取得
func (fr *FrameRange) StartFrame() float32 {
	return fr.startFrame
}

// EndFrame 終了フレームを取得
func (fr *FrameRange) EndFrame() float32 {
	return fr.endFrame
}

// Duration フレーム数を取得
func (fr *FrameRange) Duration() float32 {
	return fr.endFrame - fr.startFrame
}

// Contains 指定フレームが範囲内かどうか
func (fr *FrameRange) Contains(frame float32) bool {
	return frame >= fr.startFrame && frame <= fr.endFrame
}

// IsValid 有効な範囲かどうか
func (fr *FrameRange) IsValid() bool {
	return fr.startFrame >= 0 && fr.endFrame >= fr.startFrame
}

// String 文字列表現
func (fr *FrameRange) String() string {
	return fmt.Sprintf("[%.1f - %.1f]", fr.startFrame, fr.endFrame)
}

// Equals 他のフレーム範囲と等しいかどうか
func (fr *FrameRange) Equals(other *FrameRange) bool {
	if other == nil {
		return false
	}
	return fr.startFrame == other.startFrame && fr.endFrame == other.endFrame
}

// Overlaps 他のフレーム範囲と重複しているかどうか
func (fr *FrameRange) Overlaps(other *FrameRange) bool {
	if other == nil {
		return false
	}
	return fr.startFrame <= other.endFrame && fr.endFrame >= other.startFrame
}
