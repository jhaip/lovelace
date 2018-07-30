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

room.on(
  `camera $cameraId sees papers $papersString @ $time`,
  async ({ cameraId, papersString, time }) => {
    console.error(papersString);
    const papers = JSON.parse(papersString.replace(/'/g, '"'));
    console.error(papers);
    
    const knownPapers = (await room.select(`$processName has paper ID $paperId`))
    const visibleIDs = papers.map(paper => String(paper.id))
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
