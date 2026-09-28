package countminsketch

import "testing"

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
