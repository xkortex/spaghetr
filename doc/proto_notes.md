# Protobuf Notes

The following is some general observations I have made about the protobuf 
(and by extension gRPC) ecosystem. Assuming protobuf3

- Protoc is very, *very* particular. 
- All of the target languages of protobuf are rather opinionated about packaging,
    in myriad conflicting ways
- Protip: Arrange protobufs the way you would go files. I.e. the directory in which they
    reside is the package. 
    
##### Some code chunks to try to understand packaging patterns as they are currently 
Snippets:

`tensorflow_repo/tensorflow/core/framework/tensor.proto`

```proto
package tensorflow;
option cc_enable_arenas = true;
option java_outer_classname = "TensorProtos";
option java_multiple_files = true;
option java_package = "org.tensorflow.framework";
option go_package = "github.com/tensorflow/tensorflow/tensorflow/go/core/framework";
import "tensorflow/core/framework/resource_handle.proto";
import "tensorflow/core/framework/tensor_shape.proto";
import "tensorflow/core/framework/types.proto";

// Protocol buffer representing a tensor.
message TensorProto {
  DataType dtype = 1;

  // Shape of the tensor.  TODO(touts): sort out the 0-rank issues.
  TensorShapeProto tensor_shape = 2;
  }
  ```


`tensorflow_repo/tensorflow/core/framework/tensor_shape.proto`

```proto
option cc_enable_arenas = true;
option java_outer_classname = "TensorShapeProtos";
option java_multiple_files = true;
option java_package = "org.tensorflow.framework";
option go_package = "github.com/tensorflow/tensorflow/tensorflow/go/core/framework";

package tensorflow;

// Dimensions of a tensor.
message TensorShapeProto {}
```

### Surprisal
From these package defs, it would seem to imply that tensor_pb2.py should live 
at `from tensorflow.framework import tensor_pb2` or 
at `from tensorflow.core.framework import tensor_pb2`, latter also implied 
by the .proto file path. 

Indeed, it looks like `from tensorflow.core.framework import tensor_pb2`. 


`$ve/lib/python3.6/site-packages/tensorflow/core/framework/tensor_pb2.py`
```python
# source: tensorflow/core/framework/tensor.proto
from google.protobuf import descriptor as _descriptor


from tensorflow.core.framework import resource_handle_pb2 as tensorflow_dot_core_dot_framework_dot_resource__handle__pb2
from tensorflow.core.framework import tensor_shape_pb2 as tensorflow_dot_core_dot_framework_dot_tensor__shape__pb2
from tensorflow.core.framework import types_pb2 as tensorflow_dot_core_dot_framework_dot_types__pb2


DESCRIPTOR = _descriptor.FileDescriptor(
  name='tensorflow/core/framework/tensor.proto',
  package='tensorflow',
  syntax='proto3',)
```

`$ve/lib/python3.6/site-packages/tensorflow/core/framework/tensor_shape_pb2.py`
```python
# source: tensorflow/core/framework/tensor_shape.proto
from google.protobuf import descriptor as _descriptor
...

DESCRIPTOR = _descriptor.FileDescriptor(
  name='tensorflow/core/framework/tensor_shape.proto',
  package='tensorflow',)
```

`$ve/lib/python3.6/site-packages/tensorflow/python/eager/execute.py`
```python
from tensorflow.core.framework import tensor_pb2
from tensorflow.python.eager import core
from tensorflow.python.framework import tensor_shape

```

### Project dir
Looking at the directory `tensorflow_repo/tensorflow/core/framework/`, 
we find 
- tensor.proto
- tensor.h
- tensor.cc

However, no python or go files. However they would probably end up there. 

Let's look at some rpc files now:
`tensorflow_repo/tensorflow/python/debug/lib/debug_service_pb2_grpc.py`

```python
import grpc

from tensorflow.core.debug import debug_service_pb2 as tensorflow_dot_core_dot_debug_dot_debug__service__pb2
from tensorflow.core.protobuf import debug_pb2 as tensorflow_dot_core_dot_protobuf_dot_debug__pb2
from tensorflow.core.util import event_pb2 as tensorflow_dot_core_dot_util_dot_event__pb2
```

Hmmm

`tensorflow_repo/tensorflow/core/debug/debug_service.proto`

```proto
package tensorflow;

import "tensorflow/core/framework/tensor.proto";
import "tensorflow/core/profiler/tfprof_log.proto";
import "tensorflow/core/protobuf/debug.proto";
import "tensorflow/core/util/event.proto";
```

`$ve/lib/python3.6/site-packages/tensorflow/core/debug/debug_service_pb2.py`

```python
from tensorflow.core.framework import tensor_pb2 as tensorflow_dot_core_dot_framework_dot_tensor__pb2
```

So it seems like the proto directive `package tensorflow;` does not have much 
bearing on the compiled code. Protobuf docs say: 
>  You can add an optional package specifier to a .proto file to prevent name clashes between protocol message types.

```proto
package foo.bar;
message Open { //... 
}
```
You can then use the package specifier when defining fields of your message type:
```proto
message Foo {
  //...
  required foo.bar.Open open = 1;
  //...
}
```
> The way a package specifier affects the generated code depends on your chosen language:

> - In C++ the generated classes are wrapped inside a C++ namespace. For example, Open would be in the namespace foo::bar.
> - In Java, the package is used as the Java package, unless you explicitly provide a option java_package in your .proto file.
> - In Python, the package directive is ignored, since Python modules are organized according to their location in the file system.
> - In Go, the package directive is ignored, and the generated .pb.go file is in the package named after the corresponding go_proto_library rule.

Let's see it in action in 
`tensorflow_repo/tensorflow/core/protobuf/struct.proto`

```proto

import "tensorflow/core/framework/tensor_shape.proto";
import "tensorflow/core/framework/types.proto";

package tensorflow;
message StructuredValue {
  // The kind of value.
  oneof kind {
    // Represents None.
    NoneValue none_value = 1;

    // Represents a double-precision floating-point value (a Python `float`).
    double float64_value = 11;
    // Represents a signed integer value, limited to 64 bits.
    // Larger values from Python's arbitrary-precision integers are unsupported.
    sint64 int64_value = 12;
    // Represents a string of Unicode characters stored in a Python `str`.
    // In Python 3, this is exactly what type `str` is.
    // In Python 2, this is the UTF-8 encoding of the characters.
    // For strings with ASCII characters only (as often used in TensorFlow code)
    // there is effectively no difference between the language versions.
    // The obsolescent `unicode` type of Python 2 is not supported here.
    string string_value = 13;
    // Represents a boolean value.
    bool bool_value = 14;

    // Represents a TensorShape.
    tensorflow.TensorShapeProto tensor_shape_value = 31;
    // Represents an enum value for dtype.
    tensorflow.DataType tensor_dtype_value = 32;
    // Represents a value for tf.TensorSpec.
    TensorSpecProto tensor_spec_value = 33;
    }
}
```

Unfortunately, `python_out` protoc directives seem nowhere to be found in the 
tensorflow master repo. There is likely some crazy Bazel build abstraction
going on. Nonetheless, the files seem to end up in the right place and with the
right import statements. 



