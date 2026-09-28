# count-min-sketch

Count-Min Sketch 近似频次统计，只用 Go 标准库。

```
go test ./...
go run ./cmd/count-min-sketch --items 'a,bb,cc,dd,ee' --width 8 --depth 2
go run ./cmd/count-min-sketch --items 'a,b,a,c,a,b' --merge-halves
```

## 口径（README 为准）

- **每行要用互相独立的哈希**：`--items 'a,bb,cc,dd,ee' --width 8 --depth 2` 里 bb、cc、dd、ee 长度相同，
  现在全部落到同一列，`estimate(bb)=4`（真值 1，误差 3 超过 `e·N/width≈1.7`）；按口径每行哈希不同，
  估计值要落在 `exact ≤ estimate ≤ exact + e·N/width` 里。
- **保守更新**：Add 时只把「当前最小的那一行路径」加一，其余行不动，估计值不许无谓地虚高。
- **合并**：`Merge` 把两张同参数的表逐格相加；合并后的估计与「对合并后的输入重新建表」在同一个误差界内一致，
  并且永远不小于精确值。现在 `--merge-halves` 直接报「合并还没有实现」。
- 输入：`--width`、`--depth` 必须是正整数，否则报错。
- 不变量：估计值 ≥ 精确值；同一张表继续 Add 不会让已有条目的估计值变小；重复 Add 同一条目结果确定。
- 规模：100 万元素、1000 万次更新，内存 O(width×depth)，单次更新 O(depth)；不许保存原始条目。

## 现在的行为

- `hash` 用 FNV-1a 并把行号混进种子，不同行的落点互相独立；
  `a,bb,cc,dd,ee`（width=8）现在给 `estimate(bb)=1`。
- `Add` 保守更新：只把当前最小的那一行路径加一，其余行不动。
- `Merge` 把两张同尺寸的表逐格相加；尺寸不同会报错并说明两边尺寸。

## 命令行参数

```
--items 条目流      逗号分隔
--width / --depth   表的尺寸
--merge-halves      前后两半分别建表再合并
```

## 输出

```
estimate(bb)=1 exact(bb)=1
max_error=0 bound=1.70
merged(bb)=1
```

## 目录

```
sketch.go                建表、更新、估计与合并
cmd/count-min-sketch     命令行入口
sketch_test.go           go test 用例
```
