package trie

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

var trie Trie

func init()  {
	content, _ := ioutil.ReadFile("badkeywords_2.txt")
	str := string(content)
	split := strings.Split(str, "|")
	trie = NewTrie(split)
}

func TestTrie(t *testing.T) {
	tests := []struct {
		input    string
		output   string
		keywords []string
		found    bool
	}{
		{
			input:  "日本AV演员兼电视、电影演员。苍井空AV女优是xx出道, 日本AV女优们最精彩的表演是AV演员色情表演",
			output: "日本****兼电视、电影演员。*****女优是xx出道, ******们最精彩的表演是******表演",
			keywords: []string{
				"AV演员",
				"苍井空",
				"AV",
				"日本AV女优",
				"AV演员色情",
			},
			found: true,
		},
		{
			input:    "完全和谐的文本完全和谐的文本",
			output:   "完全和谐的文本完全和谐的文本",
			keywords: nil,
			found:    false,
		},
		{
			input:  "就一个字不对",
			output: "就*个字不对",
			keywords: []string{
				"一",
			},
			found: true,
		},
		{
			input:  "就一对, AV",
			output: "就*对, **",
			keywords: []string{
				"一",
				"AV",
			},
			found: true,
		},
		{
			input:  "就一不对, AV",
			output: "就**对, **",
			keywords: []string{
				"一",
				"一不",
				"AV",
			},
			found: true,
		},
		{
			input:  "就对, AV",
			output: "就对, **",
			keywords: []string{
				"AV",
			},
			found: true,
		},
		{
			input:  "就对, 一不",
			output: "就对, **",
			keywords: []string{
				"一",
				"一不",
			},
			found: true,
		},
		{
			input:    "",
			output:   "",
			keywords: nil,
			found:    false,
		},
	}

	trie := NewTrie([]string{
		"", // no hurts for empty keywords
		"一",
		"一不",
		"AV",
		"AV演员",
		"苍井空",
		"AV演员色情",
		"日本AV女优",
	})

	for _, test := range tests {
		strings := trie.FindKeywords(test.input)
		fmt.Println(strings)
	}
}



func BenchmarkTrie(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		trie.Filter("")
	}
}
