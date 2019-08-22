#!/usr/bin/env python
# -*- coding: utf-8 -*-

import os
from grpc_tools import protoc
import shlex
from glob import glob


def get_resource_arg():
    """Get the extra flag used when calling `python -m grpc_tools.protoc"""
    import pkg_resources
    proto_include = pkg_resources.resource_filename('grpc_tools', '_proto')
    return ['-I{}'.format(proto_include)]


def compile_protobufs(proto_path='pkgname/proto', *args):
    """compile the protobuf files.
    A few notes:
        - Madness this way lies. Seriously, protoc is insane.
        - I refuse to put `_pb2.py` files in the main namespace, as should
            every dev.
        - Protoc/protobuf is VERY TOUCHY when it comes to python, paths, and
            packages. This is likely a continuous WIP as I figure out better
            patterns and methods. This is an ongoing problem, see:
            - https://github.com/protocolbuffers/protobuf/issues/1491
            - https://github.com/grpc/grpc/issues/9575
        - I think the dot matters, sometimes.
        - Running `python -m grpc_tools.protoc` seems to have different behavior
            than calling it through this function, because of the final
            ['-I{}'.format(proto_include)]
        - If you want typical package namespacing, you are kinda stuck with
            repodir/pkgname/foobar/qux.proto which generates
            repodir/pkgname/foobar/qux_pb2.py if you want
            from pkgname.foobar import qux
        - As such, I have yet to figure out a way to leverage built-in libs
            such as google/protobuf/wrappers.proto
        - `python setup.py bdist_wheel` seems to inevitably generate the
            _pb files in the source tree. Maybe this is telling me something
        - Oh MANIFEST.in affects things here, too. What the heck.



    If you have an IDE, you might need the _pb files in place in the file tree.
    You can run `python spaghetr/compile_pb.py spaghetr/protos` in the project
    root to achieve this

    """
    use_extra=False
    file_path = os.path.join(proto_path, '{filename}')

    cmd = "--proto_path={proto_path} " \
          "--python_out={out} " \
          "--grpc_python_out={out} " \
          "{target}".format(proto_path=proto_path, out='.', target='{target}')
    filenames = glob(file_path.format(filename='*.proto'))
    print('<compile_pb> cwd      : {}'.format(os.getcwd()))
    print('<compile_pb> protopath: {}'.format(proto_path))
    print('<compile_pb> {}'.format(filenames))
    print('<compile_pb> {}'.format(cmd))

    if not filenames:
        print('<compile_pb> {}'.format("WARNING! No .proto's found to compile"))

    for fn in filenames:
        cmdf = cmd.format(target=fn)
        cmd_list = shlex.split(cmdf)
        if use_extra:
            cmd_list += get_resource_arg()
        print('<compile_pb> protoc {}'.format(' '.join(cmd_list)))

        out = protoc.main(cmd_list)
        if out:
            raise RuntimeError('Protobuf failed. Run Setup with --verbose '
                               'to see why')


if __name__ == '__main__':
    import sys
    # compile_protobufs(sys.argv[1], *sys.argv[2:])
    pkgname = 'spaghetr'
    compile_protobufs(os.path.join(pkgname, 'protos'))
