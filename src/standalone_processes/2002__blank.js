const { room, myId, run } = require('../helper2')(__filename);

/*
room.on(`$ paper 1013 is pointing at paper $editTarget`, results => {
  room.cleanup()
  let editTarget = results.length > 0 ? parseInt(results[0].editTarget) - 2000 : "nothing"
  room.assert(`wish I was labeled Editing: ${editTarget}`)
  let ill = room.newIllumination();
  ill.translate(45, 50);
  let size = 40;
  let step = size * 1.3;
  for (let x=0; x<3; x+=1) {
    for (let y=0; y<3; y+=1) {
      ill.nofill();
      if (y*3 + x == editTarget) {
        ill.fill("red");
      }
      ill.rect(x*step, y*step, size, size)
    }
  }
  room.draw(ill) 
})
*/

room.on(`$ $ keypad last button pressed is $key`, results => {
  room.cleanup()
  let key = results.length > 0 ? results[0].key : '1';
  if (key == '*' || key == '#') {
    key = '1'
  }
  let editTarget = 2000 + parseInt(key) - 1
  if (editTarget === 1999) editTarget = 1013
  room.assert(`paper 1013 is pointing at paper ${editTarget}`)
  room.assert(`wish I was labeled Editing program ${editTarget}`)
  let ill = room.newIllumination()
  ill.stroke(255, 0, 0, 128)
  ill.strokewidth(5)
  ill.nofill()
  ill.rect(2, 2, 185, 285)
  room.draw(ill, editTarget)
})


run();
