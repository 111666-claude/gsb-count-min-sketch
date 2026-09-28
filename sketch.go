// Package countminsketch 用 Count-Min Sketch 做近似频次统计，只用标准库。
package countminsketch

import "errors"

// ErrNoMerge 表示合并还没有实现。
var ErrNoMerge = errors.New("countminsketch: 合并还没有实现")

// ErrSize 表示宽度或深度不合法。
var ErrSize = errors.New("countminsketch: 宽度与深度必须为正")

// Sketch 是一个 width×depth 的计数矩阵。
type Sketch struct {
	width int
	depth int
	rows  [][]int64
}

// New 建表。缺陷：每一行共用同一个哈希，行之间没有区别。
func New(width, depth int) (*Sketch, error) {
	if width <= 0 || depth <= 0 {
		return nil, ErrSize
	}
	rows := make([][]int64, depth)
	for index := range rows {
		rows[index] = make([]int64, width)
	}
	return &Sketch{width: width, depth: depth, rows: rows}, nil
}

// hash 缺陷：只用条目长度取模，所有行落在同一列。
func (s *Sketch) hash(item string, row int) int {
	return len(item) % s.width
}

// Add 缺陷：每一行都无条件加一，没有保守更新。
func (s *Sketch) Add(item string) {
	for row := 0; row < s.depth; row++ {
		s.rows[row][s.hash(item, row)]++
	}
}

// Estimate 返回各行的最小值。
func (s *Sketch) Estimate(item string) int64 {
	best := int64(-1)
	for row := 0; row < s.depth; row++ {
		value := s.rows[row][s.hash(item, row)]
		if best < 0 || value < best {
			best = value
		}
	}
	return best
}

// Merge 缺陷：还没实现。
func (s *Sketch) Merge(other *Sketch) error {
	return ErrNoMerge
}
