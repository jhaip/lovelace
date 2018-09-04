const Room = require('@living-room/client-js')
const fs = require('fs');
const path = require('path');
const execFile = require('child_process').execFile;

const scriptName = path.basename(__filename);
const scriptNameNoExtension = path.parse(scriptName).name;
const logPath = __filename.replace(scriptName, 'logs/' + scriptNameNoExtension + ".log")
const access = fs.createWriteStream(logPath)
process.stdout.write = process.stderr.write = access.write.bind(access);
process.on('uncaughtException', function(err) {
  console.error((err && err.stack) ? err.stack : err);
})

const room = new Room()

room.subscribe(
  `wish $name would be running`,
  ({assertions, retractions}) => {
    retractions.forEach(async ({ name }) => {
      const existing_pid = await room.select(`"${name}" has process id $pid`)
      console.error(`making ${name} NOT be running`)
      console.error(existing_pid)
      existing_pid.forEach(({ pid }) => {
        pid = pid.value;
        console.log("STOPPING PID", pid)
        process.kill(pid, 'SIGTERM')
        room.retract(`"${name}" has process id $`);
        room.retract(`"${name}" is active`);
      })
    })
    assertions.forEach(async ({ name }) => {
      const existing_pid = await room.select(`"${name}" has process id $pid`)
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
            room.retract(`"${name}" has process id $`);
            room.retract(`"${name}" is active`);
            console.log(`${name} callback`)
            if (error) {
                console.error('stderr', stderr);
                console.error(error);
            }
            console.log('stdout', stdout);
        });
        const pid = child.pid;
        room.assert(`"${name}" has process id ${pid}`);
        console.error(pid);
      }
    })
  }
)
