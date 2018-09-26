const execFile = require('child_process').execFile;
const process = require('process');
const request = require('request');

const N = 10;
let nDone = 0;

const run = i => {
  const child = execFile(
    'python3',
    ['test_app.py', i],
    (error, stdout, stderr) => {
      console.log(`${i} callback`)
      if (error) {
          console.error('stderr', stderr);
          console.error(error);
      }
      console.log('stdout', stdout);
      nDone += 1;
      if (nDone === N) {
        console.timeEnd("test")
      }
  });
  const pid = child.pid;
  console.error(pid);
}

console.time("test")
for (let i = 0; i < N; i+=1) {
  run(i)
}
