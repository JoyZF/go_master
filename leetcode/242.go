package main

//
//func main()  {
//	fmt.Println(isAnagram("rbt","tar"))
//}

// 排序
//func isAnagram(s string, t string) bool {
//	s1, s2 := []byte(s), []byte(t)
//	sort.Slice(s1, func(i, j int) bool { return s1[i] < s1[j] })
//	sort.Slice(s2, func(i, j int) bool { return s2[i] < s2[j] })
//	return string(s1) == string(s2)
//}
// hash 表
func isAnagram(s string, t string) bool {
	var c1, c2 [26]int
	for _, ch := range s {
		// 实际上也是实现了一个hash function
		c1[ch-'a']++
	}
	for _, ch := range t {
		c2[ch-'a']++
	}
	return c1 == c2
}
