package pager

import (
	"github.com/gin-gonic/gin"
	"github.com/gogo/protobuf/sortkeys"
	"strconv"
	"strings"
)

type (
	// SectionKey 数据字段
	SectionKey string
	// SectionType 类型
	//  Gte: 大于等于
	//  Lte: 小于等于
	SectionType int
	// Section 查询内容
	Section map[SectionKey]map[SectionType]int64
)

const (
	// Gte 大于等于
	Gte SectionType = iota
	// Lte 小于等于
	Lte
)

func (s SectionType) String() string {
	switch s {
	case Gte:
		return "gte"
	case Lte:
		return "lte"
	}
	return ""
}

// Add 增加
func (section Section) Add(key SectionKey, typ SectionType, val int64) {
	if section[key] == nil {
		section[key] = make(map[SectionType]int64)
	}
	section[key][typ] = val
}

// Parse 解析
func (section Section) Parse(ctx *gin.Context) {
	query := ctx.Request.URL.Query()["section"]
	for _, v := range query {
		key, val := section.parseVal(v)
		switch true {
		case section.isGte(key, val):
			section.Add(section.getKey(key), Gte, section.getVal(val)[0])
		case section.isLte(key, val):
			section.Add(section.getKey(key), Lte, section.getVal(val)[0])
		case section.isGteLte(val):
			rs := section.getVal(val)
			section.Add(section.getKey(key), Gte, rs[0])
			section.Add(section.getKey(key), Lte, rs[1])
		}
	}
}

func (section Section) isGte(key, val string) bool {
	return !strings.HasPrefix(key, "-") && len(strings.Split(val, ",")) == 1
}

func (section Section) isLte(key, val string) bool {
	return strings.HasPrefix(key, "-") && len(strings.Split(val, ",")) == 1
}

func (section Section) isGteLte(val string) bool {
	return len(strings.Split(val, ",")) == 2
}

func (section Section) getVal(val string) []int64 {
	vals := strings.Split(val, ",")
	if len(vals) == 1 {
		v, err := strconv.Atoi(vals[0])
		if err != nil {
			return []int64{0}
		}
		return []int64{int64(v)}
	}

	var rs []int64
	for i := 0; i < 2; i++ {
		v, err := strconv.Atoi(vals[i])
		if err != nil {
			rs = append(rs, 0)
		} else {
			rs = append(rs, int64(v))
		}
	}
	sortkeys.Int64s(rs)
	return rs

}

func (section Section) getKey(key string) SectionKey {
	return SectionKey(strings.TrimPrefix(strings.TrimPrefix(key, "-"), "+"))
}

func (section Section) parseVal(v string) (key, val string) {
	index := strings.Index(v, ":")
	runeVal := []rune(v)
	return string(runeVal[0:index]), string(runeVal[index+1:])
}

// ParseSection 解析区间值
func ParseSection(ctx *gin.Context) Section {
	section := make(Section)
	section.Parse(ctx)
	return section
}
