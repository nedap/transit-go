package transit_go

import (
	"fmt"

	"github.com/nedap/transit-go/constants"
)

type WriteCache interface {
	CacheWrite(str string, asMapKey bool) string
	Init()
}

type writeCache struct {
	cache   map[string]string
	index   int
	enabled bool
}

func NewWriteCache(enabled bool) WriteCache {
	cache := &writeCache{enabled: enabled}
	cache.Init()
	return cache
}

func (c *writeCache) Init() {
	c.cache = make(map[string]string)
	c.index = 0
}

func (c *writeCache) CacheWrite(str string, asMapKey bool) string {
	if c.enabled && isCacheable(str, asMapKey) {
		val, found := c.cache[str]
		if found {
			return val
		} else {
			if c.index == constants.MaxCacheEntries {
				c.Init()
			}

			code := indexToCode(c.index)
			c.index += 1
			c.cache[str] = code
		}
	}
	return str
}

func indexToCode(index int) string {
	var hi int = index / constants.CacheCodeDigits
	var lo int = index % constants.CacheCodeDigits

	if hi == 0 {
		return fmt.Sprintf("%s%c", constants.SUB_STR, rune(lo+constants.BaseCharIndex))
	} else {
		return fmt.Sprintf("%s%c%c", constants.SUB_STR, rune(hi+constants.BaseCharIndex), rune(lo+constants.BaseCharIndex))
	}
}

func isCacheable(str string, asMapKey bool) bool {
	result := len(str) >= constants.MinSizeCacheable && (asMapKey || (str[0] == constants.ESC && (str[1] == ':' || str[1] == '$' || str[1] == '#')))
	return result
}
