const execFile = require('child_process').execFile;
const process = require('process');
const path = require('path');
const { room, myId, scriptName, run, MY_ID_STR } = require('../helper2')(__filename);

function runPaper(name) {
  console.error(`making ${name} be running!`)
  let languageProcess = 'node'
  let programSource = `src/standalone_processes/${name}`
  let runArgs = [programSource];
  if (name.includes('.py')) {
    languageProcess = 'python3'
  } else if (name.includes('.go')) {
    languageProcess = 'go'
    runArgs = ['run', programSource]
  }
  const child = execFile(
    languageProcess,
    runArgs,
    (error, stdout, stderr) => {
      // TODO: check if program should still be running
      // and start it again if so.
      console.error("program died:")
      console.error(name);
      console.error([["id", MY_ID_STR], ["text", name], `has process id $`])
      room.retract(["id", MY_ID_STR], ["text", name], `has process id $`);
      console.log(`${name} callback`)
      if (error) {
        console.error('stderr', stderr);
        console.error(error);
      }
      console.log('stdout', stdout);
    });
  const pid = child.pid;
  room.assert(["text", name], `has process id ${pid}`);
  console.error(pid);
}

function stopPaper(name, pid) {
  console.error(`making ${name} with PID ${pid} NOT be running`)
  process.kill(pid, 'SIGTERM')
  room.retract(`#${myId}`, ["text", name], `has process id $`);
  const dyingPaperId = (name.split(".")[0]).split("__")[0]
  console.log("done STOPPING PID", pid, "with ID", dyingPaperId)
  room.retract(`#${dyingPaperId} %`)  // clean up the dead paper's facts
}

let nameToProcessIdCache = {};

room.on(
  `$ $name has process id $pid`,
  results => {
    nameToProcessIdCache = {};
    results.forEach(result => {
      nameToProcessIdCache[result.name] = result.pid;
    })
    console.error("NEW name->PID map:")
    console.error(nameToProcessIdCache);
  }
)

room.on(
  `$ wish $name would be running`,
  results => {
    console.error("$ wish $name would be running")
    console.error(results)
    let shouldBeRunningNameToProcessIds = {};
    results.forEach(result => {
      const paperName = result.name;
      shouldBeRunningNameToProcessIds[paperName] = true;
      if (!(paperName in nameToProcessIdCache)) {
        runPaper(paperName)
      }
      // if paper already in running, let it keep running
    })
    for (var name in nameToProcessIdCache) {
      if (!(name in shouldBeRunningNameToProcessIds)) {
        stopPaper(name, nameToProcessIdCache[name])
      }
    }
  }
)

room.assert(["text", path.basename(__filename)], `has process id ${process.pid}`);

run()
