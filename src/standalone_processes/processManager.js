const Room = require('@living-room/client-js')
const execFile = require('child_process').execFile;

const room = new Room()

room.subscribe(
  `wish $name would be running`,
  ({assertions, retractions}) => {
    retractions.forEach(async ({ name }) => {
      const existing_pid = await room.select(`${name} has process id $pid`)
      console.error(`making ${name} NOT be running`)
      console.error(existing_pid)
      existing_pid.forEach(({ pid }) => {
        pid = pid.value;
        console.log("STOPPING PID", pid)
        process.kill(pid, 'SIGTERM')
        room.retract(`${name} has process id $`);
        room.retract(`${name} is active`);
      })
    })
    assertions.forEach(async ({ name }) => {
      const existing_pid = await room.select(`${name} has process id $pid`)
      if (existing_pid.length === 0) {
        console.error(`making ${name} be running!`)
        let languageProcess = 'node'
        let programSource = `src/standalone_processes/${name}`
        if (name.includes('.py')) {
          console.error("running as Python!")
          languageProcess = 'python3'
        }
        const child = execFile(
          languageProcess,
          [programSource],
          (error, stdout, stderr) => {
            // TODO: check if program should still be running
            // and start it again if so.
            room.retract(`${name} has process id $`);
            room.retract(`${name} is active`);
            console.log(`${name} callback`)
            if (error) {
                console.error('stderr', stderr);
            }
            console.log('stdout', stdout);
        });
        const pid = child.pid;
        room.assert(`${name} has process id ${pid}`);
        console.error(pid);
      }
    })
  }
)

room.assert(`processManager is active`)
room.assert('wish initialProgramCode.js would be running')
room.assert('wish printingManager.py would be running')
room.assert('wish programEditor.js would be running')
