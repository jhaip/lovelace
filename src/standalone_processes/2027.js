const { room, myId, run } = require('../helper2')(__filename);

room.cleanup()
let W = 1280
let H = 720
let ill = room.newIllumination()
ill.fill("red")
for (let x=0; x<7; x+=1) {
  for (let y=0; y<5; y+=1) {
    if ((x+y) % 2 === 0) {
      ill.rect(x*W/7, y*H/5, W/7, H/5)
    }
  }
}
room.draw(ill, "web2")




run();
