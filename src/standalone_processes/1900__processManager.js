const Room = require('@living-room/client-js')
const fs = require('fs');
const path = require('path');
const execFile = require('child_process').execFile;
const process = require('process');
const request = require('request');

const scriptName = path.basename(__filename);
const scriptNameNoExtension = path.parse(scriptName).name;
const logPath = __filename.replace(scriptName, 'logs/' + scriptNameNoExtension + ".log")
const access = fs.createWriteStream(logPath)
process.stdout.write = process.stderr.write = access.write.bind(access);
process.on('uncaughtException', function(err) {
  console.error((err && err.stack) ? err.stack : err);
})
const myId = (scriptName.split(".")[0]).split("__")[0]

const room = new Room()

const URL = 'http://localhost:3000/'

function cleanUpPaperFacts(paper) {
  console.log(`starting to clear state for #${paper}`)
  request(`${URL}facts`, { json: true }, (err, response, body) => {
    if (err) { return console.log(err); }
    const prefix = `#${paper} `
    body.assertions.forEach(a => {
      if (a.slice(0, prefix.length) === prefix) {
        room.retract(a);
      }
    })
    console.log(`done clearing state for #${paper}`)
  });
}

room.assert(`#${myId} "${path.basename(__filename)}" has process id ${process.pid}`);

room.subscribe(
  `$ wish $name would be running`,
  ({assertions, retractions}) => {
    retractions.forEach(async ({ name }) => {
      const existing_pid = await room.select(`#${myId} "${name}" has process id $pid`)
      console.error(`making ${name} NOT be running`)
      console.error(existing_pid)
      existing_pid.forEach(({ pid }) => {
        pid = pid.value;
        console.log("STOPPING PID", pid)
        process.kill(pid, 'SIGTERM')
        room.retract(`#${myId} "${name}" has process id $`);
        room.retract(`#${myId} "${name}" is active`);
        const dyingPaperId = (name.split(".")[0]).split("__")[0]
        console.log("done STOPPING PID", pid, "with ID", dyingPaperId)
        cleanUpPaperFacts(dyingPaperId)
      })
    })
    assertions.forEach(async ({ name }) => {
      const existing_pid = await room.select(`#${myId} "${name}" has process id $pid`)
      if (existing_pid.length === 0) {
        console.error(`making ${name} be running!`)
        let languageProcess = 'node'
        let programSource = `src/standalone_processes/${name}`
        let runArgs = [programSource];
        if (name.includes('.py')) {
          console.error("running as Python!")
          languageProcess = 'python3'
        } else if (name.includes('.go')) {
          console.error("running as golang")
          languageProcess = 'go'
          runArgs = ['run', programSource]
        }
        const child = execFile(
          languageProcess,
          runArgs,
          (error, stdout, stderr) => {
            // TODO: check if program should still be running
            // and start it again if so.
            room.retract(`#${myId} "${name}" has process id $`);
            room.retract(`#${myId} "${name}" is active`);
            console.log(`${name} callback`)
            if (error) {
                console.error('stderr', stderr);
                console.error(error);
            }
            console.log('stdout', stdout);
        });
        const pid = child.pid;
        room.assert(`#${myId} "${name}" has process id ${pid}`);
        console.error(pid);
      }
    })
  }
)
