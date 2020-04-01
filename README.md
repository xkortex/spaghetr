# SpaghetRPC

Wrap arbitrary subprocess call in RPC

# Building Notes

gRPC and python can clash in frustrating ways. 
Currently the best approach I've found to integrate this library with an
IDE/linter is to `pip install -e .`. This will generate the protos in the correct directory
using the `compile_pb.py` functionality.  

To reverse this, 
you can run `python setup.py develop --uninstall`, which seems wacky,
but that is the nature of legacy dependency tools. 