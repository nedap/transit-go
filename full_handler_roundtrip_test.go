package transit_go

import (
	"bytes"
	"fmt"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Custom handlers", func() {
	It("Writes nested objects using custom handlers", func() {
		type Point struct {
			x float64
			y float64
		}

		type Graph struct {
			Caption     string
			LeftPoint   Point
			CenterPoint Point
			RightPoint  Point
			Scale       float64
		}

		pointWriteHandler := WriteHandler{
			Name: "Point Write Handler",
			Tag:  func(obj interface{}) string { return "point" },
			Rep: func(obj interface{}) interface{} {
				p := obj.(Point)
				mapRepr := map[string]float64{"x": p.x, "y": p.y}
				return mapRepr
			},
		}

		graphWriteHandler := WriteHandler{
			Name: "Graph Write Handler",
			Tag:  func(obj interface{}) string { return "graph" },
			Rep: func(obj interface{}) interface{} {
				g := obj.(Graph)
				mapRepr := map[string]interface{}{
					"caption":      g.Caption,
					"left_point":   g.LeftPoint,
					"center_point": g.CenterPoint,
					"right_point":  g.RightPoint,
					"scale":        g.Scale,
				}
				return mapRepr
			},
		}

		customHandlers := map[reflect.Type]WriteHandler{
			reflect.TypeOf(Point{}): pointWriteHandler,
			reflect.TypeOf(Graph{}): graphWriteHandler,
		}

		leftPoint := Point{x: 1.1, y: 3.14}
		centerPoint := Point{x: 3.1, y: 9.2}
		rightPoint := Point{x: 6.2, y: 14.3}

		graph := Graph{
			Caption:     "My Beautiful Graph",
			LeftPoint:   leftPoint,
			CenterPoint: centerPoint,
			RightPoint:  rightPoint,
			Scale:       1.2,
		}

		var buffer bytes.Buffer
		writer := NewJSONWriterWithHandlers(&buffer, customHandlers)

		err := writer.Write(graph)
		Expect(err).To(BeNil())

		result := string(writer.Buffer().Bytes())
		Expect(result).To(MatchRegexp("\\[\"~#graph\""))

		pointReader := ReadHandler{
			Name: "Point Read Handler",
			FromRep: func(rep interface{}) (interface{}, error) {
				repAsMap, ok := rep.(map[*MapKey]interface{})
				if !ok {
					return nil, fmt.Errorf("Expected to be able to type assert to map[*MapKey]interface{}")
				}
				pointMap := make(map[string]interface{})
				for mapKey, v := range repAsMap {
					pointMap[mapKey.Key.(string)] = v
				}

				res := Point{x: pointMap["x"].(float64), y: pointMap["y"].(float64)}

				return res, nil
			},
		}

		graphReader := ReadHandler{
			Name: "Graph Read Handler",
			FromRep: func(rep interface{}) (interface{}, error) {
				repAsMap, ok := rep.(map[*MapKey]interface{})
				if !ok {
					return nil, fmt.Errorf("Expected to be able to type assert to map[*MapKey]interface{}")
				}
				graphMap := make(map[string]interface{})
				for mapKey, v := range repAsMap {
					graphMap[mapKey.Key.(string)] = v
				}

				res := Graph{
					LeftPoint:   graphMap["left_point"].(Point),
					CenterPoint: graphMap["center_point"].(Point),
					RightPoint:  graphMap["right_point"].(Point),
					Caption:     graphMap["caption"].(string),
					Scale:       graphMap["scale"].(float64),
				}

				return res, nil
			},
		}

		customReaders := ReadHandlerMap{
			"point": pointReader,
			"graph": graphReader,
		}

		readBuffer := bytes.NewBufferString(result)

		reader := NewJSONReaderWithHandlers(readBuffer, customReaders)
		readResult := reader.Read()

		resultAsGraph, ok := readResult.(Graph)
		Expect(ok)
		Expect(reflect.TypeOf(resultAsGraph)).To(Equal(reflect.TypeOf(Graph{})))
		Expect(resultAsGraph).To(Equal(graph))

	})
})
