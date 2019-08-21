#!/usr/bin/env python
# -*- coding: utf-8 -*-

import os
from grpc_tools import protoc
import shlex
from glob import glob

# todo: put this somewhere better in the file tree
def compile_protobufs(proto_path='pkgname/proto', *args):
    """compile the protobuf files.
    A few notes:
        - Madness this way lies.
        - Protoc/protobuf is VERY TOUCHY when it comes to python, paths, and
            packages. This is likely a continuous WIP as I figure out better
            patterns and methods.
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

    We want to emulate this command:
    python -m grpc_tools.protoc --proto_path=pygrpc/proto --python_out=. \
        --grpc_python_out=. pygrpc/proto/pygrpc/*.proto
    """

    # proto_path = os.path.join(dirname, protod)
    file_path = os.path.join('.', proto_path, '{filename}')

    target = './{proto_path}/time.proto'.format(proto_path=proto_path)
    cmd = "--proto_path={proto_path} " \
          "--python_out={out} " \
          "--grpc_python_out={out} " \
          "{target}".format(proto_path=proto_path, out='.', target='{target}')
    filenames = glob(file_path.format(filename='*.proto'))
    print('<compile_pb> {}'.format(proto_path))
    print('<compile_pb> {}'.format(filenames))
    print('<compile_pb> {}'.format(cmd))

    for fn in filenames:
        cmdf = cmd.format(target=fn)
        print('<compile_pb> protoc {}'.format(cmdf))

        out = protoc.main(shlex.split(cmdf))
        if out:
            raise RuntimeError('Protobuf failed. Run Setup with --verbose '
                               'to see why')


if __name__ == '__main__':
    import sys
    compile_protobufs(sys.argv[1], 'proto', *sys.argv[2:])
