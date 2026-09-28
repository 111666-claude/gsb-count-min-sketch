package countminsketch

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

func TestNewRejectsBadSize(t *testing.T) {
	if _, err := New(0, 4); err == nil {
		t.Fatal("宽度为 0 应该报错")
	}
	if _, err := New(8, 0); err == nil {
		t.Fatal("深度为 0 应该报错")
	}
}

func TestEstimateAfterOneAdd(t *testing.T) {
	sketch, err := New(8, 2)
	if err != nil {
		t.Fatalf("New 出错：%v", err)
	}
	sketch.Add("x")
	if got := sketch.Estimate("x"); got != 1 {
		t.Fatalf("加一次应该估出 1：%d", got)
	}
}

func TestEstimateUnseenIsZero(t *testing.T) {
	sketch, _ := New(8, 2)
	if got := sketch.Estimate("never"); got != 0 {
		t.Fatalf("没加过的条目应该估出 0：%d", got)
	}
}

func TestEstimateGrowsWithAdds(t *testing.T) {
	sketch, _ := New(8, 2)
	for index := 0; index < 3; index++ {
		sketch.Add("x")
	}
	if got := sketch.Estimate("x"); got != 3 {
		t.Fatalf("加三次应该估出 3：%d", got)
	}
}

func TestEstimateIsNeverBelowExact(t *testing.T) {
	sketch, _ := New(16, 3)
	for index := 0; index < 5; index++ {
		sketch.Add("abc")
	}
	sketch.Add("def")
	if got := sketch.Estimate("abc"); got < 5 {
		t.Fatalf("估计值不能小于精确值：%d", got)
	}
}

// 找两个在 row 0 冲突、row 1 不冲突的条目，供保守更新测试用。
func findCollisionPair(t *testing.T, sketch *Sketch) (string, string) {
	t.Helper()
	items := make([]string, 64)
	for index := range items {
		items[index] = fmt.Sprintf("item-%d", index)
	}
	for _, a := range items {
		for _, b := range items {
			if a != b && sketch.hash(a, 0) == sketch.hash(b, 0) &&
				sketch.hash(a, 1) != sketch.hash(b, 1) {
				return a, b
			}
		}
	}
	t.Fatal("找不到可用的冲突对")
	return "", ""
}

func TestConservativeUpdate(t *testing.T) {
	sketch, _ := New(8, 2)
	a, b := findCollisionPair(t, sketch)
	sketch.Add(a)
	sketch.Add(b)
	// b 在 row 0 与 a 冲突、row 1 不冲突：保守更新只应给 row 1 加一，
	// row 0 的冲突格必须保持 1 不动。
	if got := sketch.rows[0][sketch.hash(b, 0)]; got != 1 {
		t.Fatalf("保守更新不应给非最小行加一：row 0 冲突格=%d", got)
	}
	if got := sketch.Estimate(b); got != 1 {
		t.Fatalf("estimate(%s) 应为 1：%d", b, got)
	}
	if got := sketch.Estimate(a); got != 1 {
		t.Fatalf("estimate(%s) 应为 1：%d", a, got)
	}
}

func TestErrorBound(t *testing.T) {
	const width, depth = 64, 4
	sketch, _ := New(width, depth)
	exact := make(map[string]int64)
	var n int64
	for index := 0; index < 5000; index++ {
		item := fmt.Sprintf("key-%d", (index*7919)%300)
		sketch.Add(item)
		exact[item]++
		n++
	}
	bound := math.E * float64(n) / width
	for item, want := range exact {
		estimate := sketch.Estimate(item)
		if estimate < want {
			t.Fatalf("estimate(%s)=%d 小于精确值 %d", item, estimate, want)
		}
		if float64(estimate-want) > bound {
			t.Fatalf("estimate(%s)=%d 超出误差界 %.2f（精确值 %d）", item, estimate, bound, want)
		}
	}
}

// README 的第一个样例：width=8 depth=2 时 max_error 必须落回 e·N/width 内。
func TestReadmeSample(t *testing.T) {
	items := []string{"a", "bb", "cc", "dd", "ee"}
	sketch, _ := New(8, 2)
	for _, item := range items {
		sketch.Add(item)
	}
	bound := math.E * float64(len(items)) / 8
	var maxError int64
	for _, item := range items {
		if diff := sketch.Estimate(item) - 1; diff > maxError {
			maxError = diff
		}
	}
	if float64(maxError) > bound {
		t.Fatalf("max_error=%d 超过 bound=%.2f", maxError, bound)
	}
}

func TestMerge(t *testing.T) {
	const width, depth = 32, 3
	var items []string
	for index := 0; index < 400; index++ {
		items = append(items, fmt.Sprintf("key-%d", (index*31)%50))
	}
	left, _ := New(width, depth)
	right, _ := New(width, depth)
	rebuilt, _ := New(width, depth)
	exact := make(map[string]int64)
	half := len(items) / 2
	for _, item := range items[:half] {
		left.Add(item)
	}
	for _, item := range items[half:] {
		right.Add(item)
	}
	for _, item := range items {
		rebuilt.Add(item)
		exact[item]++
	}
	if err := left.Merge(right); err != nil {
		t.Fatalf("Merge 出错：%v", err)
	}
	bound := math.E * float64(len(items)) / width
	for item, want := range exact {
		merged := left.Estimate(item)
		if merged < want {
			t.Fatalf("合并后 estimate(%s)=%d 小于精确值 %d", item, merged, want)
		}
		if float64(merged-want) > bound {
			t.Fatalf("合并后 estimate(%s)=%d 超出误差界 %.2f", item, merged, bound)
		}
		// 与「对合并后的输入重新建表」落在同一个界内。
		if fresh := rebuilt.Estimate(item); float64(fresh-want) > bound {
			t.Fatalf("重新建表 estimate(%s)=%d 超出误差界 %.2f", item, fresh, bound)
		}
	}
	if got := left.Total(); got != int64(len(items)) {
		t.Fatalf("合并后 Total 应为 %d：%d", len(items), got)
	}
}

func TestMergeSizeMismatch(t *testing.T) {
	a, _ := New(8, 2)
	b, _ := New(16, 2)
	err := a.Merge(b)
	if err == nil {
		t.Fatal("尺寸不匹配应该报错")
	}
	if !strings.Contains(err.Error(), "8x2") || !strings.Contains(err.Error(), "16x2") {
		t.Fatalf("错误信息要说明两边尺寸：%v", err)
	}
	c, _ := New(8, 3)
	if err := a.Merge(c); err == nil {
		t.Fatal("深度不匹配也应该报错")
	}
}

func TestAddNeverDecreasesEstimate(t *testing.T) {
	sketch, _ := New(16, 3)
	sketch.Add("x")
	before := sketch.Estimate("x")
	for index := 0; index < 200; index++ {
		sketch.Add(fmt.Sprintf("noise-%d", index))
		if got := sketch.Estimate("x"); got < before {
			t.Fatalf("继续 Add 后估计值变小：%d < %d", got, before)
		}
	}
}

func TestRepeatedAddsAreDeterministic(t *testing.T) {
	build := func() *Sketch {
		sketch, _ := New(16, 3)
		for round := 0; round < 3; round++ {
			for index := 0; index < 50; index++ {
				sketch.Add(fmt.Sprintf("key-%d", index))
			}
		}
		return sketch
	}
	first, second := build(), build()
	for index := 0; index < 50; index++ {
		item := fmt.Sprintf("key-%d", index)
		if a, b := first.Estimate(item), second.Estimate(item); a != b {
			t.Fatalf("同样的输入估计值不一致：%d != %d", a, b)
		}
		if first.Estimate(item) < 0 {
			t.Fatalf("估计值不能为负：%d", first.Estimate(item))
		}
	}
}
