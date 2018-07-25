const Room = require('@living-room/client-js')
const fs = require('fs');

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
        name = name.word
        if (!name.includes('.py') && !name.includes('.js')) {
          name += '.js'
        }
        sourceCode = sourceCode.value
        room.retract(`wish ${name} has source code $`)
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
