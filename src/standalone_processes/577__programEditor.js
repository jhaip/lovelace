const Room = require('@living-room/client-js')
const fs = require('fs');
const path = require('path');

const scriptName = path.basename(__filename);
const scriptNameNoExtension = path.parse(scriptName).name;
const logPath = __filename.replace(scriptName, 'logs/' + scriptNameNoExtension + ".log")
const access = fs.createWriteStream(logPath)
process.stdout.write = process.stderr.write = access.write.bind(access);
process.on('uncaughtException', function(err) {
  console.error((err && err.stack) ? err.stack : err);
})

const room = new Room()

function main() {
  room
    .select(`wish $name has source code $sourceCode`)
    .then(data => {
      console.error(data);
      data.forEach(({ name, sourceCode }) => {
        console.error("ON WISH SOURCE CODE")
        console.error(name)
        console.error(sourceCode)
        name = name.word || name.value
        if (!name.includes('.py') && !name.includes('.js')) {
          name += '.js'
        }
        sourceCode = sourceCode.value.replace(new RegExp(String.fromCharCode(9787), 'g'), String.fromCharCode(34)) 
        console.log('debug:::')
        console.log(`wish "${name}" has source code $`)
        room.retract(`wish "${name}" has source code $`)
        fs.writeFile(`src/standalone_processes/${name}`, sourceCode, (err) => {
          if (err) throw err;
          console.error('The file has been saved!');
        });
      })
    });

  console.error("done looping in programEditor", (new Date()).getTime())
  setTimeout(main, 1000);
}

main();
