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

const readFile = readLogPath => {
  fs.readFile(readLogPath, 'utf8', (err, sourceCodeData) => {
    if (err) throw err;
    sourceCode = sourceCodeData.replace(/\n/g, '\\n').replace(/"/g, String.fromCharCode(9787))
    console.log(`"${readLogPath}" has source code "${sourceCode}"`)
    const shortFilename = path.basename(readLogPath);
    room.assert(`"${shortFilename}" has source code "${sourceCode}"`)
  });
}

const loadModulesInFolder = folder => {
  const processesFolder = path.join(__dirname, folder)
  console.log(processesFolder)
  fs.readdir(processesFolder, (_, processFiles) => {
    processFiles.forEach(processFile => {
      try {
        const processFilePath = path.join(processesFolder, processFile)
        if (!fs.lstatSync(processFilePath).isFile) return
        readFile(processFilePath)
      } catch (e) {
        console.error(e)
      }
    })
  })
}

loadModulesInFolder('.');
