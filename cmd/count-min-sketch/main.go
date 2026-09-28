// 命令行入口：对 --items 给出的条目流建 Count-Min Sketch 并打印估计。
package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"strings"

	countminsketch "example.com/countminsketch"
)

func main() {
	itemsFlag := flag.String("items", "", "逗号分隔的条目流")
	width := flag.Int("width", 0, "表的宽度（正整数）")
	depth := flag.Int("depth", 0, "表的深度（正整数）")
	mergeHalves := flag.Bool("merge-halves", false, "前后两半分别建表再合并")
	flag.Parse()

	if *width <= 0 || *depth <= 0 {
		fmt.Fprintf(os.Stderr, "错误：--width 与 --depth 必须是正整数（得到 width=%d depth=%d）\n", *width, *depth)
		os.Exit(2)
	}

	// 跳过空项。
	var items []string
	for _, item := range strings.Split(*itemsFlag, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			items = append(items, item)
		}
	}

	sketch, err := countminsketch.New(*width, *depth)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误：%v\n", err)
		os.Exit(2)
	}

	exact := make(map[string]int64)
	var order []string
	for _, item := range items {
		sketch.Add(item)
		if _, seen := exact[item]; !seen {
			order = append(order, item)
		}
		exact[item]++
	}

	var maxError int64
	for _, item := range order {
		estimate := sketch.Estimate(item)
		fmt.Printf("estimate(%s)=%d exact(%s)=%d\n", item, estimate, item, exact[item])
		if diff := estimate - exact[item]; diff > maxError {
			maxError = diff
		}
	}
	bound := math.E * float64(len(items)) / float64(*width)
	fmt.Printf("max_error=%d bound=%.2f\n", maxError, bound)

	if *mergeHalves {
		left, err := countminsketch.New(*width, *depth)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误：%v\n", err)
			os.Exit(2)
		}
		right, err := countminsketch.New(*width, *depth)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误：%v\n", err)
			os.Exit(2)
		}
		half := len(items) / 2
		for _, item := range items[:half] {
			left.Add(item)
		}
		for _, item := range items[half:] {
			right.Add(item)
		}
		if err := left.Merge(right); err != nil {
			fmt.Fprintf(os.Stderr, "错误：%v\n", err)
			os.Exit(1)
		}
		for _, item := range order {
			fmt.Printf("merged(%s)=%d\n", item, left.Estimate(item))
		}
	}
}
