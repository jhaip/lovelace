const Room = require('@living-room/client-js')
const fs = require('fs');
const path = require('path');

const scriptName = path.basename(__filename);
const scriptNameNoExtension = path.parse(scriptName).name;
const logPath = __filename.replace(scriptName, 'logs/' + scriptNameNoExtension + '.log')
const access = fs.createWriteStream(logPath)
process.stdout.write = process.stderr.write = access.write.bind(access);
process.on('uncaughtException', function(err) {
  console.error((err && err.stack) ? err.stack : err);
})

const room = new Room()

const myId = 472
const origin = [0.1, 0.1]
const lineHeight = 0.1;

console.error("HEllo from text editor")

room.subscribe(
  `paper ${myId} is pointing at paper $targetId`,
  `$targetName has paper ID $targetId`,
  `$targetName has source code $sourceCode`,
  ({assertions, retractions}) => {
    room.retract(`draw small text $ at ($, $) on paper ${myId}`)
    console.error("got stuff")
    console.error(assertions)
    console.error(retractions)
    if (retractions.length > 0) {
      room.assert(`draw small text "Point at something!" at (${origin[0]}, ${origin[1]}) on paper ${myId}`)
    }
    assertions.forEach(({targetId, targetName, sourceCode}) => {
      lines = sourceCode.split("\n")
      lines.forEach((line, i) => {
        room.assert(`draw small text "${line}" at (${origin[0]}, ${origin[1] + i * lineHeight}) on paper ${myId}`)
      })
    })
  }
)
