package main

//func main()  {
//	s := "()()()"
//	fmt.Println(isValid(s))
//}

//func isValid(s string) bool {
//	// 空字符串返回true
//	if s == "" {
//		return true
//	}
//	// 奇数字符串返回false
//	if len(s) % 2 != 0 {
//		return false
//	}
//	// 定义一个左括号的map和一个右括号的map
//	left := make(map[int]struct{},3)
//	left[40] = struct{}{}
//	left[123] = struct{}{}
//	left[91] = struct{}{}
//	right := make(map[int]struct{},3)
//	right[41] = struct{}{}
//	right[93] = struct{}{}
//	right[125] = struct{}{}
//
//	// 使用切片模拟堆操作 先进后出
//	hp := []int{}
//
//	for _,v := range s {
//		// 存在左集合放入堆
//		if _,ok := left[int(v)]; ok {
//			// 塞进堆中
//			hp = append(hp, int(v))
//		}
//		// 存在右集合 并且时配对括号
//		if _,ok := right[int(v)]; ok {
//			if len(hp)==0 {
//				return false
//			}
//			if check(hp[len(hp) -1], int(v)) {
//				hp = append(hp[0:len(hp)-1])
//			}else{
//				return false
//			}
//		}
//	}
//	// hp里元素为空时表示左右完全匹配
//	if len(hp) == 0 {
//		return true
//	}
//	return false
//}
//
//func check(i,j int) bool {
//	if i == 40 && j == 41 {
//		return true
//	}
//	if i == 91 && j == 93 {
//		return true
//	}
//	if i == 123 && j == 125 {
//		return true
//	}
//	return false
//}

func isValid(s string) bool {
	n := len(s)
	if n % 2 != 0 {
		return false
	}
	// 使用map  key时右括号 value 是左括号
	m := map[byte]byte{
		')': '(',
		'}': '{',
		']': '{',
	}
	stack := []byte{}
	for i := 0; i < n; i++ {
		if m[s[i]] > 0 { // 右括号逻辑
			// 栈元素为空 或者 左右括号不匹配 返回false
			if len(stack)  == 0 || stack[len(stack) - 1] != m[s[i]] {
				return false
			}
			stack = stack[:len(stack) - 1]
		} else {
			stack = append(stack, s[i])
		}
	}
	return len(stack) == 0
}