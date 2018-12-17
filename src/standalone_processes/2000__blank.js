const { room, myId, run } = require('../helper2')(__filename);

// Write code here!

room.cleanup()
room.assert(`Jacob says hello world`);

function drawFace(ill, current_time) {
ill.push()
ill.translate(0, 50 + 50 * Math.sin(current_time/1000.0))
// ill.fill("blue");
// ill.rect(0, 0, 200, 300);
ill.nofill()
// ill.translate(50, 50);
let face_width=100
let face_height=100
let eye_size = 10
ill.ellipse(0, 0, face_width, face_height);
ill.ellipse(face_width/4, face_height/3, eye_size, eye_size);
ill.ellipse(face_height*3/4.0, face_height/3, eye_size, eye_size);
let mouth_y = face_height*0.6
ill.line(face_width/5, mouth_y, face_width*4.0/5.0, mouth_y);
ill.pop()
}

function drawTime(ill, current_time) {
ill.fontcolor("red")
ill.text(0, 130, `time is:`)
ill.text(0, 160 ,`${current_time}`)
}

room.on(`$ time is $time`, results => {
  room.cleanup();
  let ill = room.newIllumination();
  let current_time = results.length > 0 ? results[0].time : 1;
  drawFace(ill, current_time);
  drawTime(ill, current_time);
  room.draw(ill);
})

run();