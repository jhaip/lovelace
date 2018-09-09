const Room = require('@living-room/client-js')
const fs = require('fs');
const path = require('path');
const readline = require('readline');

const scriptName = path.basename(__filename);
const scriptNameNoExtension = path.parse(scriptName).name;
const logPath = __filename.replace(scriptName, 'logs/' + scriptNameNoExtension + '.log')
const access = fs.createWriteStream(logPath)
process.stdout.write = process.stderr.write = access.write.bind(access);
process.on('uncaughtException', function(err) {
  console.error((err && err.stack) ? err.stack : err);
})
const myId = (scriptName.split(".")[0]).split("__")[0]

const room = new Room()

room.on(`$ wish everything would be printed`, async (options) => {
  room.retract(`$ wish everything would be printed`)
  const papers = (await room.select(`$ $processName has paper ID $paperId`))
  console.log("papers:")
  console.log(papers);
  papers.forEach(p => {
    const processName = p.processName.word || p.processName.value;
    room.assert(`#${myId} wish paper ${p.paperId.value} at "${processName}" would be printed`)
  })
});
