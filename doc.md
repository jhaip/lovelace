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

[Opentracing](https://opentracing.io/) & [Jaeger](https://www.jaegertracing.io/) - useful for debugging and tracing during low level development

## Building up the software of a room

Every space is different and should be built to fit the needs, experience, and goals of the people that work within it. The room is not an app you download and run, but something that should be built by you. But spaces do have similar baseline needs:
* A way to control what programs as running. This usually means some way for humans to understand what is running and some way for sensors to understand what should be running to match that.
* Some sort of display that running programs can use for visual feedback
* A way to edit programs and create new programs

####  An example room:

Programs are represented physically as papers with code written on them. When a paper is face-up, showing the code, then it is running. Projectors in the ceiling project graphics on the programs as the running programs instruct them to. To edit and create programs, there is a special piece of paper that edits the code of whatever paper it is pointing at. The code being edited is projected on to the text editor paper and a wireless keyboard associated with the text editor edits the text. When a new version of code is saved, a new version of the paper is printed out of a new sheet of paper to replace the old piece of paper. For the room to notice what papers are face-up, papers are marked with patterns of colored dots in their colors for identification. Cameras in the ceiling looking for these dots to figure out where papers are how coordinate with the projector to project the right graphics onto the right paper.

Paper sensing:
1. Camera
2. Grab a frame from the webcam and find everything that looks like a dot
3. Get the list of dots and figure out what papers they map to using the knowledge that papers have only certains patterns of dots in the four corners of a paper
4. Get the list of visual papers and wish that the corresponding program stored on a computer was running.
5. Get all wishes for programs that should be running and run them on a computer.

Program editing:
1. The input from a wireless keyboard is captured and claimed to the room
2. On boot, a program reads the contents of all code files and claims them to the room
3. The text editor paper:
    a. Gets the source of the program it is editing
    b. Listens for the latest key presses to control the text editor cursor
    c. Wishes the text editor graphics would be projected on it
4. When saving a program, the text editor wishes a program would be saved
5. A program lists for wishes of edited programs, transforms the code from the room's domain specific language into, and then wishes that the files on disk would be edited and run.
6. Simultaneously when saving a program, another program generates a PDF of a new piece of paper and then another program talks to the printer to print the PDF.
7. A computer's files on disk are used as a local way to persist the contents of files and to send them to accesories like printers but this is an implementation choice and not something that someone in the room would see or need to care about.

Projection:
1. A process listens for the locations of all papers and each papers wishes for graphics to be drawn on them. Additionally it listens for a projector-camera calibration so the display is able to perform the projection mapping.

### Alternative inputs

Instead of sensing papers with cameras looking for colored dots, put a RFID card of the back of every paper and embedded RFID sensors in the room. A RFID sensor detecting a card is the equilalent of "the paper is in the room" and that is should be running.

### Alternative outputs

Sounds, Stands of light, physical movement of things in the rooom, smells.
