from helper2 import init, claim, retract, prehook, subscription, batch, check_server_connection
import helper2
import logging
import time
import os
import sys
import os, signal

my_pid = int(os.getpid())

def check_kill_process(pstring):
    for line in os.popen("ps ax | grep " + pstring + " | grep -v grep"):
        fields = line.split()
        pid = fields[0]
        if int(pid) == my_pid:
            print("It's my own PID!")
        else:
            print("not my PID {}".format(pid))
        os.kill(int(pid), signal.SIGKILL)

helper2.rpc_url = "10.0.0.22"

if len(sys.argv) != 2:
    print("Expected a single argument of the process to run!")

process_name = sys.argv[1]

@prehook
def my_prehook():
    # Kill process
    print("killing process")
    # print("pkill -f \"{}\"".format(process_name))
    # os.system("pkill -f \"{}\"".format(process_name))
    check_kill_process(process_name)
    # Restart process
    print("starting new process: python3 {} &".format(process_name))
    os.system("python3 {} &".format(process_name))

init(__file__, skipListening=True)

while True:
    time.sleep(10)
    print("checking server connection")
    check_server_connection()
