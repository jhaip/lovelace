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
let fontSize = 32;
let fontHeight = fontSize / 1080.0;
let lineHeight = 1.3 * fontHeight;
const origin = [0.0001, 0.0001 + lineHeight]
let charWidth = fontHeight * 0.38;
const cursorColor = `(255, 128, 2)`
let cursorPosition = [10, 10]
let editorWidthCharacters = 40
let editorHeightCharacters = 20

console.error("HEllo from text editor")

room.subscribe(
  `paper ${myId} is pointing at paper $targetId`,
  `$targetName has paper ID $targetId`,
  `$targetName has source code $sourceCode`,
  ({assertions, retractions}) => {
    room.retract(`draw $ text $ at ($, $) on paper ${myId}`)
    room.retract(`draw a ${cursorColor} line from ($, $) to ($, $) on paper ${myId}`)
    console.error("got stuff")
    console.error(assertions)
    console.error(retractions)
    if (retractions.length > 0) {
      room.assert(`draw "${fontSize}pt" text "Point at something!" at (${origin[0]}, ${origin[1]}) on paper ${myId}`)
    }
    assertions.forEach(({targetId, targetName, sourceCode}) => {
      lines = sourceCode.split("\n")
      console.error(lines)
      lines.slice(0, editorHeightCharacters).forEach((lineRaw, i) => {
        const line = lineRaw.substring(0, editorWidthCharacters);
        room.assert(`draw "${fontSize}pt" text "${line}" at (${origin[0]}, ${origin[1] + i * lineHeight}) on paper ${myId}`)
      });
      room.assert(
        `draw a ${cursorColor} line from ` +
        `(${origin[0] + cursorPosition[0] * charWidth}, ${origin[1] + cursorPosition[1] * lineHeight})` +
        ` to ` +
        `(${origin[0] + cursorPosition[0] * charWidth}, ${origin[1] + cursorPosition[1] * lineHeight - fontHeight})` +
        ` on paper ${myId}`
      );
    })
  }
)
