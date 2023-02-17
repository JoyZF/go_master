package main

func replaceSpace(s string) string {
	space := " "
	bytes := []byte(s)
	res := []byte{}
	for _, v := range bytes {
		if string(v) == space {
			res = append(res, []byte("%20")...)
		} else {
			res = append(res, v)
		}
	}
	return string(res)
}
