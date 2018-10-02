N = 100
i = N

def ping():
    pong()

def pong():
    global i
    i -= 1
    if i > 0:
        ping()

import time
start = time.time()
ping()
end = time.time()
print("TIME for N={}: {} us".format(N, (end - start)*1000.0*1000.0))
