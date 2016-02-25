package transit_go

import (
	"fmt"

	"github.com/jwkoelewijn/transit-go/constants"
)

type baseParser struct {
	readHandlerMap ReadHandlerMap
	defaultHandler *DefaultReadHandler
	mapBuilder     MapReader
	arrayBuilder   ArrayReader
	parser         Parser
}

func (p *baseParser) parseString(str string) (interface{}, error) {
	if len(str) > 1 {
		switch str[0] {
		case constants.ESC:
			switch str[1] {
			case constants.ESC, constants.SUB, constants.RESERVED:
				return str[1:len(str)], nil
			case constants.TAG:
				tag := Tag(str[2:len(str)])
				return tag, nil
			default:
				return p.decode(str[1:2], str[2:len(str)])
			}
		case constants.SUB:
			if str[1] == ' ' {
				return constants.MAP_AS_ARRAY, nil
			}
		}
	}
	return str, nil
}

func (p *baseParser) decode(tag string, rep interface{}) (interface{}, error) {
	handler, err := p.readHandlerMap.lookupHandler(tag)
	if err == nil {
		readHandler, ok := handler.(ReadHandler)
		if !ok {
			return nil, fmt.Errorf("Could not decode %s (%s) because the handler is not a ReadHandler", tag, rep)
		}
		return readHandler.FromRep(rep.(string))
	} else if p.defaultHandler != nil {
		return p.defaultHandler.FromRep(tag, rep)
	} else {
		return nil, fmt.Errorf("Cannot fromRep %s: %+v", tag, rep)
	}
}
