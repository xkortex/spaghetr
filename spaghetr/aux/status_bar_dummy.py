#!/usr/bin/env python

import sys
import time
import argparse

F1 = sys.stdout
F2 = sys.stderr


def arg_parser():
    import argparse
    parser = argparse.ArgumentParser(description="""
    Status bar go brrrrr
    """)
    parser.add_argument(
        '-l', '--log', action='store_true',
        help="Emulate a log output")
    parser.add_argument(
        '-q', '--cumulative', action='store_true',
        help="Delay between events cumulatively increases")
    parser.add_argument(
        '-c', '--code', action='store', type=int, default=0,
        help="System exit code")
    parser.add_argument(
        '-d', '--delay', action='store', type=float, default=0.05,
        help="Delay between events")
    parser.add_argument(
        '-n', '--num', action='store', type=int, default=20,
        help="Number of iterations")
    parser.add_argument(
        'input', nargs='*', type=str,
        help="Some random input")
    return parser


def dummy_bar(n=20, delay=0.05, cumulative=False):
    for i in range(n):
        s = '[{: <20}]'.format('#' * i)
        F1.write(s + '\r')
        F1.flush()
        time.sleep(0.1)
        if i == 5:
            F2.write(' ' * len(s) + '\r')
            F2.write('oops\n\r')
            F2.flush()
            time.sleep(delay * (i*cumulative or 1))


def dummy_logs(n=20, delay=0.05, cumulative=False):
    for i in range(n):

        s = '[{: >3}][{:}] '.format(i, time.time())
        if i & 1:
            F1.write(s + str(F1) + ' \n')
            F1.flush()
        else:
            F2.write(s + str(F2) + ' \n')
            F2.flush()
        time.sleep(delay * (i*cumulative or 1))


if __name__ == '__main__':
    args = arg_parser().parse_args()
    print(args)

    if args.log:
        dummy_logs(args.num, args.delay, args.cumulative)
        exit(args.code)

    dummy_bar(args.num, args.delay, args.cumulative)
    exit(args.code)
