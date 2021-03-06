#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""Receives gRPC calls and converts them to subprocess calls"""


import time
import socket
import subprocess as subp
from concurrent import futures

import grpc
from vprint import vprint

from spaghetr.protos import status_pb2, status_pb2_grpc
from spaghetr.protos import basic_subproc_pb2, basic_subproc_pb2_grpc
from spaghetr.util import parse_nullable

DEFAULT_PORT = 45654

class Status(status_pb2_grpc.StatusServicer):
    def GetStatus(self, request, context):
        vprint('Got request at {}'.format(time.ctime()))
        hostip = socket.gethostbyname(socket.gethostname())
        return status_pb2.StatusReply(message=time.ctime(), ip=hostip,
                              fqdn=socket.getfqdn(),
                              host=socket.gethostname())


class Subprocesser(basic_subproc_pb2_grpc.SubprocessServicer):

    def Run(self, request: basic_subproc_pb2.ArgsRequest,
              context) -> basic_subproc_pb2.StdReply:
        vprint('Got request at {}'.format(time.ctime()))
        result = basic_subproc_pb2.StdReply()
        p = subp.run(request.args, input=parse_nullable(request.input),
                     stdout=subp.PIPE, stderr=subp.PIPE)
        result.stdout = p.stdout
        result.stderr = p.stderr
        result.returncode = p.returncode
        return result

    def Popen(self, request: basic_subproc_pb2.ArgsRequest,
              context) -> basic_subproc_pb2.StdReply:
        vprint('Got request at {}'.format(time.ctime()))
        result = basic_subproc_pb2.StdReply()
        p = subp.Popen(request.args)
        result.stdout, result.stderr = p.communicate()
        result.pid = p.pid
        result.returncode = p.returncode
        return result


def arg_parser():
    import argparse
    parser = argparse.ArgumentParser(description="""Run subprocess as RPC""")

    parser.add_argument(
        "-H", "--host", default=None, action="store", type=str,
        help="Host to run RPC service on")
    parser.add_argument(
        "-P", "--port", default=DEFAULT_PORT, action="store", type=str,
        help="start of port range to run RPC service on")
    return parser


def serve(host=None, port=DEFAULT_PORT):
    if host is None:
        host = '[::]'

    hostport = ':'.join([host, str(port)])

    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    status_pb2_grpc.add_StatusServicer_to_server(Status(), server)
    basic_subproc_pb2_grpc.add_SubprocessServicer_to_server(
        Subprocesser(), server)
    server.add_insecure_port(hostport)
    vprint('starting server on {}'.format(hostport))
    server.start()
    try:
        while True:
            time.sleep(3600)
    except KeyboardInterrupt:
        server.stop(0)


if __name__ == '__main__':
    args = arg_parser().parse_args()
    serve(args.host, args.port)
