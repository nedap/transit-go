Attempt to implement Cognitect's Transit format in Go.
See http://transit-format.org

Implementation 
==============
The implementation is a translation from transit-java and follows the same principles. Some of them could probably be simplified or be made more Go'ish.

At the moment the ReadHandlers are implemented as structs, because this enables factory methods to actually construct them.
This choice was made because in Go it is not possible to create anonymous interface implementations and it did not feel right
to create dozens of types just to use an interface. So now ReadHandler's are a struct with a FromRep member, which has a type
of func(rep interface{}) interface{}.


Compatibility
=============

At the moment only the JSON reader and writer have been implemented and they have not been tested against the roundtrip tests that were
released as part of the Transit specification.

Because of the typeless nature of the Transit format, the implementation can only return interface{} types, so when you want
to create custom ReadHandler's for your own type, you have to do the casting and type assertions yourself. This is also the
reason that Array's will be decoded as []interface{} and maps as map[interface{}]interface{}. This does mean that the standard
roundtrip tests need some thinking.

Note that this is still an early implementation, so expect to find bugs.

Future work
===========

JSONVerbose and MessagePack Reader and Writer implementations are not implemented yet.

This implementation should be tested against the test-set provided with the specification.

The code could be structured a bit better :)
