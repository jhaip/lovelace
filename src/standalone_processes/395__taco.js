const Room = require('@living-room/client-js')
const fs = require('fs');
const path = require('path');

const scriptName = path.basename(__filename);
const scriptNameNoExtension = path.parse(scriptName).name;
const logPath = __filename.replace(scriptName, 'logs/' + scriptNameNoExtension + ".log")
const access = fs.createWriteStream(logPath)
process.stdout.write = process.stderr.write = access.write.bind(access);
process.on('uncaughtException', function(err) {
  console.error((err && err.stack) ? err.stack : err);
})

const room = new Room()

console.log("start taco")
room.retract(`hello from taco @ $`)

setInterval(() => {
  console.error("hello from taco", new Date())
  room
    .retract(`hello from taco @ $`)
    .assert(`hello from taco @ ${(new Date()).getTime()}`)
}, 1000);