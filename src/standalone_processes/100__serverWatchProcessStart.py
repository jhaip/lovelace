from helper2 import init, claim, retract, prehook, subscription, batch, check_server_connection
import helper2
import logging
import time
import os
import sys

helper2.rpc_url = "10.0.0.22"

if len(sys.argv) != 2:
    print("Expected a single argument of the process to run!")

process_name = sys.argv[1]

@prehook
def my_prehook():
    # Kill process
    print("killing process")
    print("killall -9 \"{}\"".format(process_name))
    os.system("killall -9 \"{}\"".format(process_name))
    # Restart process
    print("starting new process: {} &".format(process_name))
    os.system("{} &".format(process_name))

init(__file__, skipListening=True)

while True:
    time.sleep(10)
    print("checking server connection")
    check_server_connection()
