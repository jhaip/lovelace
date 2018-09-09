const fs = require('fs');
const { room, myId } = require('../helper')(__filename);

function main() {
  room
    .select(`$wisherId wish $name has source code $sourceCode`)
    .then(data => {
      console.error(data);
      data.forEach(({ name, sourceCode, wisherId }) => {
        console.error("ON WISH SOURCE CODE")
        console.error(name)
        console.error(sourceCode)
        console.error(wisherId)
        name = name.word || name.value
        if (!name.includes('.py') && !name.includes('.js')) {
          name += '.js'
        }
        sourceCode = sourceCode.value.replace(new RegExp(String.fromCharCode(9787), 'g'), String.fromCharCode(34))
        wisherId = wisherId.id
        console.log('debug:::')
        console.log(`#${wisherId} wish "${name}" has source code $`)
        room.retract(`#${wisherId} wish "${name}" has source code $`)
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
