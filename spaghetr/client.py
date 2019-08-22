import sys
import socket
import shlex

import grpc
from vprint import vprint

from spaghetr.protos import status_pb2, status_pb2_grpc
from spaghetr.protos import basic_subproc_pb2, basic_subproc_pb2_grpc
from spaghetr.util import parse_nullable


HOST = 'localhost'
PORT = 45654


def run_client(host=None, port=PORT, cmd=None):
    if host is None:
        # host = socket.gethostname()
        host = 'localhost'

    if cmd is None:
        cmd = ['ls']
    else:
        cmd = shlex.split(cmd)

    hostport = ':'.join([host, str(port)])
    channel = grpc.insecure_channel(hostport)
    stub = basic_subproc_pb2_grpc.SubprocessStub(channel)
    msg = basic_subproc_pb2.ArgsRequest()
    for c in cmd:
        msg.args.append(c)
    response = stub.Run(msg)
    # print('Client received: {}'.format(response.message))
    return str(response)


def run_status(host=None, port=PORT, cmd=None):
    if host is None:
        # host = socket.gethostname()
        host = 'localhost'

    if cmd is None:
        cmd = ['ls']

    hostport = ':'.join([host, str(port)])
    channel = grpc.insecure_channel(hostport)
    stub = status_pb2_grpc.StatusStub(channel)
    response = stub.GetStatus(status_pb2.NullRequest())
    # print('Client received: {}'.format(response.message))
    return str(response)


def arg_parser():
    from argparse import ArgumentParser
    parser = ArgumentParser()

    parser.add_argument(
        "-H", "--host", default=None, action="store", type=str,
        help="Host to run RPC service on")
    parser.add_argument(
        "-P", "--port", default=PORT, action="store", type=str,
        help="start of port range to run RPC service on")
    parser.add_argument(
        "-+", "--health", action="store_true",
        help="Run a health check")
    parser.add_argument(
        'cmd', nargs='?', default='ls', type=str,
        help="Command to run"
    )

    return parser


if __name__ == '__main__':
    vprint('vprint on {}'.format(socket.gethostname()))
    args = arg_parser().parse_args()
    vprint('{}'.format(args))

    if args.health:
        out = run_status(args.host, args.port)
        print(out)
        sys.exit(0)

    out = run_client(args.host, args.port, args.cmd)
    print(out)



