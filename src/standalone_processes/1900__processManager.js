const Room = require('@living-room/client-js')
const execFile = require('child_process').execFile;

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

room.assert('wish "390__initialProgramCode.js" would be running')
room.assert('wish "498__printingManager.py" would be running')
room.assert('wish "577__programEditor.js" would be running')
room.assert('wish "826__runSeenPapers.js" would be running')
room.assert('wish "277__pointingAt.py" would be running')
room.assert('wish "620__paperDetails.js" would be running')

room.assert(`camera 1 has projector calibration TL (0, 0) TR (1920, 0) BR (1920, 1080) BL (0, 1080) @ 1`)

// room.assert(`camera 1 sees paper 1924 at TL (100, 100) TR (1200, 100) BR (1200, 800) BL (100, 800) @ 1`)
// room.assert(`paper 1924 is pointing at paper 472`)  // comment out if pointingAt.py is running
