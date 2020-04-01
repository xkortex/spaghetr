import sys
import time

if __name__ == '__main__':
    for i in range(20):
        s = '[{: <20}]'.format('#' * i)
        print(s, end='\r', flush=True)
        time.sleep(0.1)
        if i == 5:
            sys.stderr.write(' ' * len(s) + '\r')
            sys.stderr.write('oops\n\r')
            sys.stderr.flush()
            time.sleep(0.05)
