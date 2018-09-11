const fs = require('fs');
const path = require('path')
const { room, myId, scriptName } = require('../helper')(__filename);

const readFile = readLogPath => {
  try {
    sourceCodeData = fs.readFileSync(readLogPath, 'utf8');
    sourceCode = sourceCodeData.replace(/\n/g, '\\n').replace(/"/g, String.fromCharCode(9787))
    // console.log(`"${readLogPath}" has source code "${sourceCode}"`)

    const shortFilename = path.basename(readLogPath);
    let paperId = "";
    if (!shortFilename.includes(".")) {
      console.log("skipping the binary", shortFilename)
      return;
    }
    if (shortFilename === '__init__.py' || shortFilename === 'helper.py') {
      return;
    }
    if (shortFilename.includes("__")) {
      paperId = shortFilename.split("__")[0];
    } else if (shortFilename.includes(".")) {
      paperId = shortFilename.split(".")[0];
    }
    console.log(`#${myId} "${shortFilename}" has paper ID ${paperId}`)

    room.assert(`#${myId} "${shortFilename}" has source code "${sourceCode}"`)
    room.assert(`#${myId} "${shortFilename}" has paper ID ${paperId}`)
    console.log(`done with "${shortFilename}"`)
  } catch (e) {
    console.error("readLogPath", readLogPath)
    console.error(e);
  }
}

const loadModulesInFolder = folder => {
  const processesFolder = path.join(__dirname, folder)
  console.log(processesFolder)
  const processFiles = fs.readdirSync(processesFolder);
  console.log(processFiles);
  console.log("---")
  processFiles.forEach(processFile => {
    try {
      const processFilePath = path.join(processesFolder, processFile)
      console.log(fs.lstatSync(processFilePath).isFile())
      if (!fs.lstatSync(processFilePath).isFile()) return
      readFile(processFilePath)
    } catch (e) {
      console.error(e)
    }
  })
}

loadModulesInFolder('.');
// TODO: remove this HACK
// need a way to exit the node program after all promises have returned
// promises come from asserting and retracting things in the lovelace library code
setTimeout(() => {
  process.exit()
}, 5000)
