#!/usr/bin/env python

import sys
import time

F1 = sys.stdout
F2 = sys.stderr

if __name__ == '__main__':
    for i in range(20):
        s = '[{: <20}]'.format('#' * i)
        F1.write(s + '\r')
        F1.flush()
        time.sleep(0.1)
        if i == 5:
            F2.write(' ' * len(s) + '\r')
            F2.write('oops\n\r')
            F2.flush()
            time.sleep(0.05)
