// Package sensitivewords 敏感词过滤
//  example:
//  words := sensitivewords.NewSensitiveWords().LoadKeywords(app.Root() + "/configs/debug/keywords.csv")
//	fmt.Println(words.HasKeywords("鸡巴"))
package sensitivewords

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"unicode"
)

type node struct {
	isEnd bool
	next  Store
}

// Store 关键字节点
type Store map[rune]*node

type keywords struct {
	skip           int
	mappings       map[rune][]int
	sensitiveWords [][]rune
}

// 特殊符号
var symbol = regexp.MustCompile("\\s|~|`|!|@|#|\\$|%|\\^|&|\\*|\\(|\\)|-|_|=|\\+|\\[|]|;|:|'|\"|/|\\?|\\.|>|,|<")
var file *os.File

func (keywords *keywords) Strings() []string {
	var words = make([]string, 0)
	for i := 0; i < len(keywords.sensitiveWords); i++ {
		words = append(words, string(keywords.sensitiveWords[i]))
	}

	return words
}

func (keywords *keywords) replace(str string) string {
	var newStr = []rune(str)
	for _, indexes := range keywords.mappings {
		for _, index := range indexes {
			newStr[index] = '*'
		}
	}
	return string(newStr)
}

// NewSensitiveWords 创建一个对象来加载关键字
func NewSensitiveWords() Store {
	return make(Store)
}

// LoadKeywords 加载文件中的关键字, 按行分割
func (store Store) LoadKeywords(filename string) Store {
	var err error
	file, err = os.OpenFile(filename, os.O_RDWR|os.O_APPEND, os.ModePerm)
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(file)
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}

		store.Add(string(line))
	}

	return store
}

// AppendToFile 增加新的关键字时, 将同时写入到文件
func (store Store) AppendToFile(keywords string) (n int, err error) {
	store.Add(keywords)
	return file.WriteString(fmt.Sprintf("%s\n", keywords))
}

// Add 增加关键字
func (store Store) Add(keywords string) {
	words := []rune(keywords)
	length := len(words)

	if length == 0 {
		return
	}

	if _, ok := store[words[0]]; !ok {
		store[words[0]] = &node{isEnd: length == 1, next: NewSensitiveWords()}
	}
	if !store[words[0]].isEnd {
		store[words[0]].isEnd = length == 1
	}
	store[words[0]].next.Add(string(words[1:]))
}

// 内部查找敏感词
func (store Store) find(input string, triggered *keywords) {
	words := []rune(input)
	nowNode := store

	tmpWord := make([]rune, 0)
	for i := 0; i < len(words); i++ {
		current := words[i]

		// 特殊符号和数字跳过
		if symbol.MatchString(string(current)) || unicode.IsDigit(current) {
			continue
		}

		// 非字母和中文跳过
		if !unicode.IsLetter(current) && !unicode.Is(unicode.Scripts["Han"], current) {
			continue
		}

		val, nodeExists := nowNode[current]
		if !nodeExists {
			for _, tmp := range tmpWord {
				delete(triggered.mappings, tmp)
			}
			tmpWord = make([]rune, 0)
			nowNode = store
			val, nodeExists = nowNode[current]
		}

		if nodeExists {
			if val.isEnd {
				triggered.mappings[current] = append(triggered.mappings[current], i)
				tmpWord = append(tmpWord, current)
				triggered.sensitiveWords = append(triggered.sensitiveWords, tmpWord)
				nowNode = store
				tmpWord = make([]rune, 0)
			} else {
				triggered.mappings[current] = append(triggered.mappings[current], i)
				tmpWord = append(tmpWord, current)
				nowNode = val.next
			}
		}
	}
}

// HasKeywords 是否包含敏感词
func (store Store) HasKeywords(input string) bool {
	var words = &keywords{mappings: make(map[rune][]int), sensitiveWords: make([][]rune, 0)}
	store.find(input, words)
	return len(words.sensitiveWords) > 0
}

// KeywordsList 获取触发的敏感词列表
func (store Store) KeywordsList(input string) []string {
	var words = &keywords{mappings: make(map[rune][]int), sensitiveWords: make([][]rune, 0)}
	store.find(input, words)
	return words.Strings()
}

// Filter 过滤敏感词
func (store Store) Filter(input string, excludes ...string) string {
	var words = &keywords{mappings: make(map[rune][]int), sensitiveWords: make([][]rune, 0)}
	store.find(input, words)
	return words.replace(input)
}
