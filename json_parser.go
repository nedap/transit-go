package transit_go

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nedap/transit-go/constants"
)

type JsonParser struct {
	decoder  *json.Decoder
	base     baseParser
	curToken *json.Token
}

func NewJsonParser(decoder *json.Decoder, handlers ReadHandlerMap, defaultHandler *DefaultReadHandler, mapBuilder MapReader, listBuilder ArrayReader) Parser {
	jsonParser := JsonParser{decoder: decoder}

	baseParser := baseParser{
		readHandlerMap: handlers,
		defaultHandler: defaultHandler,
		mapBuilder:     mapBuilder,
		arrayBuilder:   listBuilder,
		parser:         jsonParser,
	}
	jsonParser.base = baseParser
	return jsonParser
}

func (p JsonParser) currentToken() json.Token {
	return *p.curToken
}

func (p *JsonParser) nextToken() json.Token {
	if token, err := p.decoder.Token(); err == nil {
		p.curToken = &token
	} else {
		p.curToken = nil
	}
	return *p.curToken
}

func (p JsonParser) parseString(str string) (interface{}, error) {
	return p.base.parseString(str)
}

func (p JsonParser) parseLong() (interface{}, error) {
	var long int64
	err := p.decoder.Decode(&long)
	return long, err
}

func (p JsonParser) parse(cache ReadCache) (interface{}, error) {
	if !p.decoder.More() {
		return nil, fmt.Errorf("Unexpected EOF")
	} else {
		p.nextToken()
		return p.parseVal(false, cache)
	}
}

func (p JsonParser) parseVal(asMapKey bool, cache ReadCache) (interface{}, error) {
	token := p.currentToken()
	if token == nil {
		return nil, nil
	}
	if delim, ok := token.(json.Delim); ok {
		switch delim {
		case '{':
			return p.parseMap(asMapKey, cache, nil)
		case '[':
			return p.parseArray(asMapKey, cache, nil)
		}
	} else if str, ok := token.(string); ok {
		return cache.CacheRead(str, asMapKey, p), nil
	} else if b, ok := token.(bool); ok {
		return b, nil
	} else if num, ok := token.(json.Number); ok {
		if strings.Contains(num.String(), ".") {
			return num.Float64()
		} else {
			bigInt, err := num.Int64()
			if err != nil {
				return nil, nil
			}
			return int(bigInt), nil
		}
	} else if token == nil {
		return nil, nil
	}
	return nil, nil
}

func (p JsonParser) parseMap(asMapKey bool, cache ReadCache, handler *MapReadHandler) (interface{}, error) {
	return p.parseMapUntilToken(asMapKey, cache, handler, '}')
}

func (p JsonParser) parseMapUntilToken(asMapKey bool, cache ReadCache, handler *MapReadHandler, endRune rune) (interface{}, error) {
	var mr MapReader
	if handler == nil {
		mr = p.base.mapBuilder
	} else {
		mr = handler.mapReader
	}

	mb := mr.Init()

	nextToken := p.nextToken()

	for !tokenEquals(nextToken, string(endRune)) {

		key, err := p.parseVal(true, cache)
		if err != nil {
			return nil, err
		}

		if tag, ok := key.(Tag); ok {
			valHandler, err := p.base.readHandlerMap.lookupHandler(string(tag))

			var val interface{}
			// advance to read value
			token := p.nextToken()
			if token != nil {
				mapHandler, isMapHandler := valHandler.(MapReadHandler)
				arrayHandler, isArrayHandler := valHandler.(ArrayReadHandler)
				if token == "{" && isMapHandler {
					val, err = p.parseMap(false, cache, &mapHandler)
					if err != nil {
						return nil, err
					}
				} else if token == "[" && isArrayHandler {
					val, err = p.parseArray(false, cache, &arrayHandler)
					if err != nil {
						return nil, err
					}
				} else {
					parsedVal, err := p.parseVal(false, cache)
					if err != nil {
						return nil, err
					}
					readHandler, ok := valHandler.(ReadHandler)
					if !ok {
						return nil, err
					}
					val, err = readHandler.FromRep(parsedVal)
					if err != nil {
						return nil, err
					}
				}
			} else {
				parsedVal, err := p.parseVal(false, cache)
				if err != nil {
					return nil, err
				}
				val, err = p.base.decode(string(tag), parsedVal)
				if err != nil {
					return nil, err
				}
			}
			// advance to read end of array or object
			p.nextToken()
			return val, nil
		} else {
			p.nextToken()
			val, err := p.parseVal(false, cache)
			if err != nil {
				return nil, err
			}
			mb = mr.Add(mb, key, val)
		}

		nextToken = p.nextToken()
	}

	return mr.Complete(mb), nil
}

func (p JsonParser) parseArray(ignored bool, cache ReadCache, handler *ArrayReadHandler) (interface{}, error) {
	nextToken := p.nextToken()

	if nextToken != "]" {
		firstVal, err := p.parseVal(false, cache)
		if err != nil {
			return nil, err
		}
		if firstVal != nil {
			tagTag, isTag := firstVal.(Tag)

			if firstVal == constants.MAP_AS_ARRAY {
				// if the same, build a map with rest array contents
				return p.parseMapUntilToken(false, cache, nil, ']')
			} else if isTag {
				tag := string(tagTag)
				valHandler, err := p.base.readHandlerMap.lookupHandler(tag)

				var val interface{}
				if err == nil {
					mapHandler, isMapHandler := valHandler.(MapReadHandler)
					arrayHandler, isArrayHandler := valHandler.(ArrayReadHandler)

					currentToken := p.nextToken()
					if err != nil {
						return nil, err
					}
					tokenDelim, isDelim := currentToken.(json.Delim)
					currentTokenString := string(tokenDelim)
					if isDelim && currentTokenString == "{" && isMapHandler {
						val, err = p.parseMap(false, cache, &mapHandler)
						if err != nil {
							return nil, err
						}
					} else if isDelim && currentTokenString == "[" && isArrayHandler {
						val, err = p.parseArray(false, cache, &arrayHandler)
						if err != nil {
							return nil, err
						}
					} else {
						// read value and decode normally
						handler, _ := valHandler.(ReadHandler)
						parsedVal, err := p.parseVal(false, cache)
						if err != nil {
							return nil, err
						}
						val, err = handler.FromRep(parsedVal)
						if err != nil {
							return nil, err
						}
					}
				} else {
					// default decode
					parsedVal, err := p.parseVal(false, cache)
					if err != nil {
						return nil, err
					}
					val, err = p.base.decode(tag, parsedVal)
				}
				// advance past the end of the object or array
				p.nextToken()
				return val, nil
			}
		}

		// process array without special decoding or interpretation
		var arrayReader ArrayReader
		if handler != nil {
			arrayReader = handler.arrayReader
		} else {
			arrayReader = p.base.arrayBuilder
		}

		ab := arrayReader.Init(0)
		ab = arrayReader.Add(ab, firstVal)
		p.nextToken()
		for !tokenEquals(p.currentToken(), "]") {
			nextVal, err := p.parseVal(false, cache)
			if err == nil {
				ab = arrayReader.Add(ab, nextVal)
			}
			p.nextToken()
		}
		completeArray := arrayReader.Complete(ab)
		return completeArray, nil
	}
	// Make an empty collection, using handler's array reader, if present
	var arrayReader ArrayReader
	if handler != nil {
		arrayReader = handler.arrayReader
	} else {
		arrayReader = p.base.arrayBuilder
	}
	return arrayReader.Complete(arrayReader.Init(0)), nil
}

func tokenEquals(token json.Token, str string) bool {
	delim, isDelim := token.(json.Delim)
	if isDelim {
		return string(delim) == str
	}
	return false
}
