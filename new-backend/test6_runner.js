const execFile = require('child_process').execFile;
const process = require('process');
const request = require('request');

const N = 12;
let nDone = 0;
const M = 3;
let mDone = 0;

const runServers = i => {
  const child = execFile(
    'python3',
    ['test6_server.py', (N/M)*i, (N/M)*(i+1)-1],
    (error, stdout, stderr) => {
      console.log(`server ${i} callback`)
      if (error) {
          console.error('stderr', stderr);
          console.error(error);
      }
      console.log('stdout', stdout);
      mDone += 1;
      if (mDone === M) {
        console.timeEnd("test")
      }
  });
  const pid = child.pid;
  console.error(pid);
}

const runClients = i => {
  const child = execFile(
    'python3',
    ['test6_client.py', i],
    (error, stdout, stderr) => {
      console.log(`client ${i} callback`)
      if (error) {
          console.error('stderr', stderr);
          console.error(error);
      }
      console.log('stdout', stdout);
      nDone += 1;
      if (nDone === N) {
        console.timeEnd("test")
        process.exit(1);
      }
  });
  const pid = child.pid;
  console.error(pid);
}

console.time("test")
for (let i = 0; i < M; i+=1) {
  runServers(i)
}
for (let i = 0; i < N; i+=1) {
  runClients(i)
}
