# Room

## Bill of Materials

1. A computer. Preferrably running linux but I have also done some work on macos. It could be a cloud server but for things that happen locally to a room it makes a little more sense to keep the networking local.
    * Currently I'm using a Intel NUC mini PC with 16 GB RAM, a quad-core Intel i5 at 1.30GHz, and a SSD. 16 GB RAM seems to be excessive in my tests but having multiple cores are important when many programs are running.

2. For camera sensing of papers: a webcam. I use the [Logitech C922](https://www.logitech.com/en-us/product/c922-pro-stream-webcam) to get HD video at 30fps.

3. For RFID sensing of papers: RFID Reader, RFID cards, and something that can read from the sensor and make networking requests.
    * I got several of these [RFID sensor + card packs](https://www.microcenter.com/product/476359/rfid-read-and-write-module). Generally any RC522 sensor and 13.56MHz cards should work.
    * I used [Particle Photon](https://www.adafruit.com/product/2721) boards. They are small, have built in WiFi, support over-the-air firmware updates, and have a nice ecosystem.
    * A [breadboard](https://www.adafruit.com/product/640) + [wires](https://www.adafruit.com/product/153) are also needed for wiring.
    * I made a "code stand" out of a [clear plastic photo frame](https://www.dollartree.com/special-moments-freestanding-borderless-vertical-plastic-photo-frames-8x10-in/225471) and a [metal chopper](https://www.dollartree.com/cooking-concepts-stainless-steel-chopper-scrapers/226405) from Dollar Tree.
    * [9x12 in clipboards from Dollar Tree](https://www.dollartree.com/wooden-clipboards-9x12-in/188681)

4. When using a projector as a display: Any projector should work but HD projectors are nice when projecting small text. Projectors with higher lumens will make the projected graphics more visible in daylight. I used the [Epson Home Cinema 1060](https://www.amazon.com/Epson-Cinema-brightness-speakers-projector/dp/B073S4TS4G/) when projecting across the room and the [Optima GT1080 short throw projector](https://www.amazon.com/Optoma-GT1080Darbee-Lumens-Gaming-Projector/dp/B06XHG92Y5/) when projecting down at a coffee table or the ground.

5. A generic color 2D printer

6. A computer monitor or TV as a supplemental display.
    * A monitor is useful when starting the computer and when debugging
    * I use a TV as a supplemental display. Best used to display one thing as a time in full screen, such as a text editor. TVs and monitors are less immersive than projection, but they can be useful when used as a pure display and not as the screen for a computer's operating system.

## Software Requirements

Operating system:
* Tested with Ubuntu 16.04
* Partial tests with MacOS High Sierra

Node.js:
* [Node.js v10.6.0](https://nodejs.org/en/download/package-manager/)
* `npm install`

Python:
* [Python 3.5+](https://www.python.org/downloads/)
* `pip3 install -r requirements.txt`
* [wxPython4](https://wxpython.org/)
* [pyGame 1.9.5+](https://www.pygame.org)
* [OpenCV 3.4](https://opencv.org/)

Golang:
* [Golang 1.12](https://golang.org/)
* Dependencies are tracked in the `go.mod` file and will be automatically downloaded and installed when running a Go program for the first time.

Particle Photon:
* [Particle Photon Workbench](https://www.particle.io/workbench/)
* A [Particle.io account](https://login.particle.io/signup) with your Particle Photon boards assocaited to your account.
* Particle Photon libraries:
    * [HttpClient](https://github.com/nmattisson/httpclient)
    * SparkJson

Printing Papers:
* [`lpr`](http://man7.org/linux/man-pages/man1/lpr.1.html) with a default printer configured.

Opentracing?
