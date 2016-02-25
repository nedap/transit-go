package constants

const (
	ESC          = '~'
	ESC_STR      = "~"
	TAG          = '#'
	TAG_STR      = "#"
	SUB          = '^'
	SUB_STR      = "^"
	RESERVED     = '`'
	ESC_TAG      = "~#"
	QUOTE_TAG    = "~#'"
	MAP_AS_ARRAY = "^ "

	MinSizeCacheable = 4
	CacheCodeDigits  = 44
	MaxCacheEntries  = CacheCodeDigits * CacheCodeDigits
	BaseCharIndex    = 48
)
