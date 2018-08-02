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

const room = new Room()

room.subscribe(
  `camera $cameraId sees paper $id at TL ($x1, $y1) TR ($x2, $y2) BR ($x3, $y3) BL ($x4, $y4) @ $time`,
  async ({ assertions }) => {
    console.log(assertions);
    const knownPapers = (await room.select(`$processName has paper ID $paperId`))
    const visibleIDs = assertions.map(paper => String(paper.id))
    console.log("knownPapers", knownPapers)
    console.log("visibleIDs", visibleIDs)

    knownPapers.forEach(paper => {
        const processName = paper.processName.word;
        if (visibleIDs.includes(String(paper.paperId.value))) {
          room.assert(`wish ${processName} would be running`);
        } else {
          room.retract(`wish ${processName} would be running`);
        }
    });
  }
)
