const fs = require('fs');
const { room, myId } = require('../helper')(__filename);

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
