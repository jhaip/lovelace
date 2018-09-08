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

room.subscribe(
  `$ camera $cameraId sees paper $id at TL ($x1, $y1) TR ($x2, $y2) BR ($x3, $y3) BL ($x4, $y4) @ $time`,
  async ({ assertions }) => {
    if (!assertions || assertions.length === 0) {
      return;
    }
    console.log("ASSERTIONS:")
    console.log(assertions);
    const knownPapers = (await room.select(`$ $processName has paper ID $paperId`))
    const visibleIDs = assertions.map(paper => String(paper.id))
    console.log("knownPapers", knownPapers)
    console.log("visibleIDs", visibleIDs)
    const bootPapers = ["0", "498", "577", "826", "277", "620", "1459", "1800", "1382", "1900", "989"]

    knownPapers.forEach(paper => {
        const processName = paper.processName.word || paper.processName.value;
        if (visibleIDs.includes(String(paper.paperId.value))) {
          console.error(`wish "${processName}" would be running`)
          room.assert(`#${myId} wish "${processName}" would be running`);
        } else if (!bootPapers.includes(String(paper.paperId.value))) {
          console.error(`RETRACT: wish "${processName}" would be running`)
          room.retract(`#${myId} wish "${processName}" would be running`);
        }
    });
  }
)

room.subscribe(
  `$ camera 1 sees no papers @ $time`,
  async ({ assertions }) => {
    if (!assertions || assertions.length === 0) {
      return;
    }
    console.log("no papers, stopping all programs")
    const knownPapers = (await room.select(`$ $processName has paper ID $paperId`))
    knownPapers.forEach(paper => {
        const processName = paper.processName.word;
        room.retract(`#${myId} wish "${processName}" would be running`);
    });
  }
)
