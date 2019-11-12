[![CircleCI](https://circleci.com/gh/nedap/transit-go.svg?style=svg)](https://circleci.com/gh/nedap/transit-go)

Attempt to implement Cognitect's Transit format in Go.
See http://transit-format.org

# Example

```go
import (
  "bytes"
  "fmt"
  . "github.com/nedap/transit-go"
)

func main() {
  thing := map[int]string{1: "hello", 2: "world"}

  var buffer bytes.Buffer
  writer := NewJSONWriter(&buffer)
  err := writer.Write(thing)

  if err != nil {
    panic(err)
  }
  str := string(writer.Buffer().Bytes())
  fmt.Println(str)
  // Outputs: ["^ ","~i1","hello","~i2","world"]
  reader := NewJSONReader(&buffer)
  result := reader.Read()

  fmt.Printf("%+v\n", result)
  // Outputs: map[1:hello 2:world]
  // Mind that the return type of Read() is interface{} and the map is of type map[interface{}]interface{}
}
```

# Implementation

The implementation is a translation from transit-java and follows the same principles. Some of them could probably be simplified or be made more Go'ish.

At the moment the ReadHandlers are implemented as structs, because this enables factory methods to actually construct them.
This choice was made because in Go it is not possible to create anonymous interface implementations and it did not feel right
to create dozens of types just to use an interface. So now ReadHandler's are a struct with a FromRep member, which has a type
of func(rep interface{}) interface{}.

To be able to create maps of all types, it should be possible to use arrays or other maps as keys as well (because this is possible
to encode in Transit). Because Go does not allow slices or maps to be used as keys in maps, whenever a map or slice is encountered as key,
this is decoded by using a pointer to a MapKey struct (because pointers _are_ allowed as keys). This trick is really just to get around this limitation.

Arrays are in fact allowed to be used as keys of a map, but the Go compiler will not allow you to create an array of a size determined by
the value of a variable, even though the creation will only happen once, and the value is known when memory should be allocated for it.
This really forced us to use the MapKey abomination. I am open for any suggestions to get around this, and I will welcome PR's with open arms :)

# Compatibility

At the moment only the JSON reader and writer have been implemented and they have not been tested against the roundtrip tests that were
released as part of the Transit specification.

Because of the typeless nature of the Transit format, the implementation can only return interface{} types, so when you want
to create custom ReadHandler's for your own type, you have to do the casting and type assertions yourself. This is also the
reason that Array's will be decoded as []interface{} and maps as map[interface{}]interface{}. This does mean that the standard
roundtrip tests need some thinking.

Note that this is still an early implementation, so expect to find bugs.

# Future work

JSONVerbose and MessagePack Reader and Writer implementations are not implemented yet.

This implementation should be tested against the test-set provided with the specification.

The code could be structured a bit better :)
