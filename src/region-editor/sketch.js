var longPollingActive = true;
var SCALE_FACTOR = 6;

let sketchMaker = function (regionData, regionPointChanged) {
  let sketch = function (p) {
    // console.log(p)
    let CANVAS_WIDTH = 1920 / SCALE_FACTOR;
    let CANVAS_HEIGHT = 1080 / SCALE_FACTOR;
    let w = 10;
    let h = 10;
    let locations = [
      [regionData.x1, regionData.y1],
      [regionData.x2, regionData.y2],
      [regionData.x3, regionData.y3],
      [regionData.x4, regionData.y4]];
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
          regionPointChanged(regionData.id, locations);
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
  return sketch;
};

var lastLoggedRegionTime = (new Date()).getTime();
var THROTTLE_TIME_MS = 1000 * 0.5;

function regionPointChanged(regionId, regionPoints) {
  let currentTime = (new Date()).getTime();
  if (currentTime - lastLoggedRegionTime > THROTTLE_TIME_MS) {
    lastLoggedRegionTime = currentTime;
    fetch(`/region/${regionId}`, {
      method: 'put',
      headers: {
        'Content-Type': 'application/json'
        // 'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: JSON.stringify({
        'x1': regionPoints[0][0],
        'y1': regionPoints[0][1],
        'x2': regionPoints[1][0],
        'y2': regionPoints[1][1],
        'x3': regionPoints[2][0],
        'y3': regionPoints[2][1],
        'x4': regionPoints[3][0],
        'y4': regionPoints[3][1]
      })
    }).then(
      response => console.log(response)
    );
  }
}

function makeRegion(data) {
  let $parent = document.createElement('div');
  $parent.setAttribute("class", "parent");
  let $leftCol = document.createElement('div');
  $leftCol.setAttribute("class", "left-col");
  let $el = document.createElement('div');
  $el.setAttribute("id", data.id);
  let $id = document.createElement('small')
  $id.innerHTML = `${data.id}`;
  let $name = document.createElement('input')
  $name.setAttribute("type", "text")
  $name.setAttribute("placeholder", "Region name")
  $name.value = `${data.name}`;
  let $toggleableGroup = document.createElement('div');
  let $toggleable = document.createElement('input')
  $toggleable.setAttribute("type", "checkbox");
  $toggleable.checked = data.toggleable;
  let $toggleLabel = document.createElement('label');
  $toggleLabel.innerHTML = 'Toggeable region?'
  $toggleableGroup.appendChild($toggleable);
  $toggleableGroup.appendChild($toggleLabel);
  let $deleteButton = document.createElement('button');
  $deleteButton.innerHTML = 'Delete Region'
  $leftCol.appendChild($id);
  $leftCol.appendChild($name);
  $leftCol.appendChild($toggleableGroup);
  $leftCol.appendChild($deleteButton);
  $parent.appendChild($el);
  $parent.appendChild($leftCol);
  document.body.appendChild($parent);
  $deleteButton.onclick = (evt) => {
    evt.preventDefault();
    fetch(`/region/${data.id}`, {
      method: 'delete'
    }).then(
      response => console.log(response)
    );
  }
  $toggleable.onchanged = (evt) => {
    fetch(`/region/${data.id}`, {
      method: 'put',
      headers: {
        'Content-Type': 'application/json'
        // 'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: JSON.stringify({
        'toggleable': $toggleable.checked
      })
    }).then(
      response => console.log(response)
    );
  }
  let myp5_c1 = new p5(sketchMaker(data, regionPointChanged), data.id);
}

function update(data) {
  document.body.innerHTML = '';
  data.forEach(datum => {
    makeRegion({
      'id': datum.id,
      'name': datum.name,
      'x1': datum.x1 / SCALE_FACTOR,
      'y1': datum.y1 / SCALE_FACTOR,
      'x2': datum.x2 / SCALE_FACTOR,
      'y2': datum.y2 / SCALE_FACTOR,
      'x3': datum.x3 / SCALE_FACTOR,
      'y3': datum.y3 / SCALE_FACTOR,
      'x4': datum.x4 / SCALE_FACTOR,
      'y4': datum.y4 / SCALE_FACTOR,
      'toggleable': true
    });
  })
  // makeRegion({
  //   'id': '9df78dc0-9e97-4a63-851e-b5bd61ba55c6',
  //   'name': 'pl1health',
  //   'x1': 20,
  //   'y1': 20,
  //   'x2': 100,
  //   'y2': 20,
  //   'x3': 100,
  //   'y3': 100,
  //   'x4': 20,
  //   'y4': 100,
  //   'toggleable': true
  // });
}

// update([]);

async function loop() {
  try {
    const response = await fetch('/status')
    const myJson = await response.json();
    update(myJson);
    if (longPollingActive) {
      setTimeout(function () {
        loop();
      }, 1000);
    }
  } catch (error) {
    console.error(error);
  }
}

loop();
