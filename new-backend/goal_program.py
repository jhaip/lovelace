from helper import *

N = 100
i = 0
start = None

def prehook():
    claim("Bird has 5 toes")

@when("$ $X {} has $Y toes".format(MY_ID))
def f(results):
    global i, N, start
    # print("sub CALLBACK!")
    # print(results)
    # print(i)
    i += 1
    if i >= N:
        end = time.time()
        print("TIME: {} ms".format((end - start)*1000.0))
        print("AVG LOOP TIME: {} ms".format((end - start)*1000.0/N))
        sys.exit()
    claim("Man {} has {} toes".format(MY_ID, i))

run()
