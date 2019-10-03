let sketch = function (p) {
  // console.log(p)
  let SCALE_FACTOR = 6;
  let CANVAS_WIDTH = 1920 / SCALE_FACTOR;
  let CANVAS_HEIGHT = 1080 / SCALE_FACTOR;
  let w = 10;
  let h = 10;
  let locations = [
    [CANVAS_WIDTH * 0.2, CANVAS_HEIGHT * 0.2],
    [CANVAS_WIDTH * 0.8, CANVAS_HEIGHT * 0.2],
    [CANVAS_WIDTH * 0.8, CANVAS_HEIGHT * 0.8],
    [CANVAS_WIDTH * 0.2, CANVAS_HEIGHT * 0.8]];
  let draggingIndex = -1;
  let offsetX, offsetY;
  let indexToCornerName = {0: "TL", 1: "TR", 2: "BR", 3: "BL"}

  p.setup = function () {
    p.createCanvas(CANVAS_WIDTH, CANVAS_HEIGHT);
  };

  p.draw = function () {
    p.background(200);

    p.beginShape();
    for (let i=0; i<4; i+=1) {
      if (i === draggingIndex) {
        locations[i][0] = p.mouseX + offsetX;
        locations[i][1] = p.mouseY + offsetY;
      }
      p.noStroke();
      if (i === draggingIndex) {
        p.fill(0, 0, 0, 150);
      } else {
        p.fill(128, 0, 0, 150);
      }
      p.rect(locations[i][0], locations[i][1], w, h);
      p.fill(0, 0, 100);
      p.text(indexToCornerName[i], locations[i][0], locations[i][1] - 5);
      p.vertex(locations[i][0], locations[i][1]);
    }
    p.noFill();
    p.stroke(100);
    p.endShape(p.CLOSE);
  };

  p.mousePressed = function () {
    if (draggingIndex !== -1) {
      return;
    }
    for (let i = 0; i < 4; i += 1) {
      let x = locations[i][0];
      let y = locations[i][1];
      if (p.mouseX > x && p.mouseX < x + w && p.mouseY > y && p.mouseY < y + h) {
        draggingIndex = i;
        offsetX = x - p.mouseX;
        offsetY = y - p.mouseY;
      }
    }
  }

  p.mouseReleased = function () {
    draggingIndex = -1;
  }
};

function makeRegion(data) {
  let el = document.createElement('div');
  el.setAttribute("id", data.id);
  document.body.appendChild(el)
  let myp5_c1 = new p5(sketch, data.id);
}

makeRegion({'id': 'a63ee29c-f5b3-4b8b-a91f-93f1a7151c06'});