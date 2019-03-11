const { room, myId, run } = require('../helper2')(__filename);

const code = `
void setup() {
  pinMode(D7, OUTPUT);
}

void loop() {
  delay(1000);
  digitalWrite(D7, HIGH);
  delay(1000);
  digitalWrite(D7, LOW);
}
`
room.assert(`wish`, ["text", code], `runs on the photon`)




run();
