# #1 Pure Python Ping Pong N=100:
100 us - 200 us
1-2 us per iteration

# #2 Pure Go Ping Pong N=100:
800 ns - 1.2 us (1200 ns)
8 ns - 12 ns per iteration
- Golang is ~100x faster than #1

# #3 Pure C Ping Pong N=1000:
6 us = 6000 ns
6 ns or less per iteration
- Golang is maybe twice as slow?

# #4 Python zmq ping pong N=1000:
512 ms
0.512 ms (512 us) per iteration
~300x times slower than Pure python in #1
- 51ms for 100 iterations, already too slow for 60fps (16ms)

# #5 Golang zmq ping pong, N=1000:
450 ms
0.45 ms (450 us) per iteration

# Other: hypothetical IPC zmq push/pull in Python
10.37 ms for N=100
0.1ms (100 us) per iteration

# Full Python, parse, SQLite, zmq, N=100:
333 ms
3.33 ms (3330 us) per iteration
- 6.5x as slow as pure python in zmq

# Full Golang, parse, sqlite, zmq, N=100:
300ms
3 ms (3000 us) per iteration
- put together query results: 429us - 800us

# Full Golang, parse, sqlite, zmq, subscriber notification is 1 query, N=100:
170ms
1.7ms per iteration
- time to read query results is still the slowest part at up to 1.5ms per iteration
- Assuming 10 iterations, this is in budget for 60fps!
- In golang, it would probably be faster to not use SQLite but there may be
advantages for concurrency
- Critical path meets budget but this needs to be scaled horizontally to 1000 concurrent paths

# #8 Nanomsg, PAIR, tcp, N=100:
19ms
190 us per iteration
- less than half the time to iterate for zmq pub/sub
- could be PAIR, or it could be nanomsg

-----

budget
16 ms = 16000 us to do 100 transactions
0.16 ms = 160 us per iteration
For python zmq ping pong (0.5ms): already too slow
60 us leftover for work
