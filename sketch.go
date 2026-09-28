// Package countminsketch 用 Count-Min Sketch 做近似频次统计，只用标准库。
package countminsketch

import (
	"errors"
	"fmt"
	"math"
)

// ErrSize 表示宽度或深度不合法。
var ErrSize = errors.New("countminsketch: 宽度与深度必须为正")

// Sketch 是一个 width×depth 的计数矩阵。
// 只保存计数矩阵，不保存原始条目：内存 O(width×depth)。
type Sketch struct {
	width int
	depth int
	rows  [][]int64
	total int64
}

// New 建表。宽度与深度都必须是正整数。
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

// FNV-1a 的参数（64 位）。
const (
	fnvOffsetBasis = 14695981039346656037
	fnvPrime       = 1099511628211
)

// hash 用 FNV-1a 打散条目，并把行号混进种子，
// 保证不同行的落点互相独立。
func (s *Sketch) hash(item string, row int) int {
	h := uint64(fnvOffsetBasis) ^ (uint64(row+1) * fnvPrime)
	for i := 0; i < len(item); i++ {
		h ^= uint64(item[i])
		h *= fnvPrime
	}
	return int(h % uint64(s.width))
}

// Add 保守更新：先算出各行的当前最小值，
// 只把等于最小值的那些行路径加一，其余行不动。
// 单次更新 O(depth)。
func (s *Sketch) Add(item string) {
	s.total++
	minValue := int64(math.MaxInt64)
	for row := 0; row < s.depth; row++ {
		if value := s.rows[row][s.hash(item, row)]; value < minValue {
			minValue = value
		}
	}
	for row := 0; row < s.depth; row++ {
		column := s.hash(item, row)
		if s.rows[row][column] == minValue {
			s.rows[row][column]++
		}
	}
}

// Estimate 返回各行的最小值，永不为负。
func (s *Sketch) Estimate(item string) int64 {
	best := int64(math.MaxInt64)
	for row := 0; row < s.depth; row++ {
		if value := s.rows[row][s.hash(item, row)]; value < best {
			best = value
		}
	}
	return best
}

// Total 返回累计的更新次数。
func (s *Sketch) Total() int64 {
	return s.total
}

// Merge 把另一张同尺寸的表逐格加进本表；尺寸不同则报错并说明两边尺寸。
func (s *Sketch) Merge(other *Sketch) error {
	if other == nil {
		return fmt.Errorf("countminsketch: 不能合并空表（本表 %dx%d）", s.width, s.depth)
	}
	if s.width != other.width || s.depth != other.depth {
		return fmt.Errorf("countminsketch: 尺寸不匹配，不能合并：本表 %dx%d，对方 %dx%d",
			s.width, s.depth, other.width, other.depth)
	}
	for row := 0; row < s.depth; row++ {
		for column := 0; column < s.width; column++ {
			s.rows[row][column] += other.rows[row][column]
		}
	}
	s.total += other.total
	return nil
}
