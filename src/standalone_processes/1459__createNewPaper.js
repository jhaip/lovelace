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

// const dotCodes = fs.readFileSync("./mytext.txt", "utf-8").split("\n");
room.on(
  `$wisherId wish a paper would be created in $language with source code $sourceCode @ $time`,
  async ({ wisherId, language, sourceCode, time }) => {
    // choose ID that is unique
    const existingIds = (await room.select(`$ $ has paper ID $id`)).map(p => p.id.value);
    console.error("Existing IDs")
    console.error(existingIds);
    let newId = null;
    while (newId === null || existingIds.includes(newId)) {
      newId = Math.floor(Math.random() * 8400/4)
    }
    console.log("new id", newId);

    // create a new file with the source code
    cleanSourceCode = sourceCode.replace(new RegExp(String.fromCharCode(9787), 'g'), String.fromCharCode(34))
    const shortFilename = `${newId}.${language}`;
    fs.writeFile(`src/standalone_processes/${shortFilename}`, cleanSourceCode, (err) => {
      if (err) {
        return console.log(err);
      }
      sourceCodeNewlineCleaned = sourceCode.replace(/\n/g, '\\n')
      room.retract(`#${wisherId.id} wish a paper would be created in ${language} with source code $ @ ${time}`)
      room.assert(`#${myId} "${shortFilename}" has source code "${sourceCodeNewlineCleaned}"`)
      room.assert(`#${myId} "${shortFilename}" has paper ID ${newId}`)
      room.assert(`#${myId} wish paper ${newId} at "${shortFilename}" would be printed`)
    });
  }
);
