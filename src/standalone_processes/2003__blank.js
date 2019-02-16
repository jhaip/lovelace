const { room, myId, run } = require('../helper2')(__filename);

room.on(`$ time is $time`, results => {
  room.cleanup()
  let ill = room.newIllumination()
  ill.translate(0, 50)
  let size = 200
  ill.nostroke()
  ill.ellipse(0, 0, size, size)
  ill.stroke("white")
  ill.translate(size/2, size/2)
  let time = results.length > 0 ? parseInt(results[0].time) : 0
  let minute = (time/(1000*60)) % 60
  let hour = (time/(1000*60*60)) % 12
  let hourAngle = -(hour/12.0)*2*Math.PI + Math.PI/2.0
  let minuteAngle = -(minute/60.0)*2*Math.PI + Math.PI/2.0
  let secondAngle = -(((time/1000)%60)/60.0)*2*Math.PI + Math.PI/2.0
  ill.strokewidth(10)
  ill.line(0, 0, size/3 * Math.cos(hourAngle), -size/3 * Math.sin(hourAngle))
  ill.strokewidth(5)
  ill.line(0, 0, size/2.2 * Math.cos(minuteAngle), -size/2.2 * Math.sin(minuteAngle))
  ill.strokewidth(1)
  ill.line(0, 0, size/2 * Math.cos(secondAngle), -size/2 * Math.sin(secondAngle))
  room.draw(ill)
})

run();