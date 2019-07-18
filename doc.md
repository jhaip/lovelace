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

## Core idea: A Shared Fact Table

Programs communicate by making and subscribing to changes in a "shared fact table". Facts are sentences like "Temperature is 23 C", which is parsed in the fact table as a typed list of phrases like [(text, Temperature), (text, is), (int, 23), (text, C)]. "Claiming a fact" or "asserting a fact" means adding something new to the shared fact table. "Retracting" or "clearing" a fact means removing zero or more things from the shared fact table. Programs subscribe to fill-in-the-blank questions about the shared fact table such as "When Temperature is $X". `$X` represents a variable that is filled in when the subscribed program is notified about a new changed result set. Subscriptions can have multiple parts and reuse variables. For example "When $program is at $x $y, $program has source code $code" would return distict results of x, y, program, and code where the value of `program` is the same across the query parts. This query style is similar to [Datalog](https://en.wikipedia.org/wiki/Datalog) and it draws inspiration from HARC's [Natural Language Datalog](https://github.com/harc/nl-datalog). The shared fact table is used as a communication mechanism between programs for this project because it:
* Allows the people writing programs to claims facts and subscriptions that read closer to a sentence.
* Allows programs to communicate asyncronously without needing to know about other programs.
* Maps well to the idea that programs in the same physical space share common knowledge.

When a program is stopped, all the facts and subscriptions is made in the shared fact table are removed. This keeps the shared fact table relevant only the programs that are currently visible in the room and running. Additionally all facts in the shared fact table record what program claimed the fact so the same fact claimed by two different programs is considered two unique facts.

The idea of a shared fact table can be thought of as a "database" or a "global state" but I don't think they are good descriptions. The shared fact table is similar to database in that changes happen via additions or deletions and there is a query langauge. The shared fact table is different in that the data is limited to one list of sentence-like facts rather than tables, columns, nodes, or partitions. The limited type of data allows the query language is to also resemble a sentence without needing to most of the features of a query language like SQL. Global state and it's associated negative connotation is also not a good description of the shared fact table because of the fact's relation to the physical program that claimed them. If a program is stopped (by being physically removed from the space) then it's facts are immediately removed. In this way, the global state is not polluted by unknown programs. Also each program operates in isolation and much explicitly subscribe to facts in the global state. The same sentence claimed by two separate programs would still be saved as two separate facts.

The shared fact table is managed by a "server". All programs in the room are aware of a server and communicate with the server to make changes to the fact table and to hear about changes to their subscriptions about the shared fact table. Currently programs and the server communicate via [ZeroMQ messages](http://zeromq.org/). The server should "local" to the programs in the room and only serve about as many programs that can fit in a single "space" or "room". For larger spaces or to connect multiple rooms together, facts can be explicitly shared to other servers. In this way data and programs are federated. Federation is useful both technically (because a single server cannot hold all the facts in the universe) and to aid in the understanding of the people working in a space. For example the position of a paper on a table is very important to someone reading the paper at the same table, but the position of a paper in a different room is irrelevant unless the person at the table as some special reason they care about something going on somewhere else.

## Building up the software of a room

Every space is different and should be built to fit the needs, experience, and goals of the people that work within it. The room is not an app you download and run, but something that should be built by you. But spaces do have similar baseline needs:
* A way for humans and computers to understand what programs are running and which are not.
* Some sort of display that running programs can use for visual feedback
* A way to edit programs and create new programs

After deciding on a way people will know what programs and running and which are not, an appropriate sensor can be chosen to detect that. For example if a program is represented by a piece of paper with code on it, then a running program could be when it is face-up and visible to everyone in the room. A sensor that could be used to detect face-up papers is a camera.

It is convenient to think of computer programs as a mirror of the physical objects they represent. If a piece of paper has code on it and should be running, then a program on some computer should have a program with the same source code and should be running on that computer. In this way programs on a computer aren't started or stopped by interacting with a traditional computer desktop with a mouse, but by changing something physically in the room. The computer that is running the programs to match the physical world are just an implementation detail and not something that people in the space should need to think about.

#### An example room:

Programs are represented physically as papers with code written on them. When a paper is face-up, showing the code, then it is running. A camera is used as a sensor. Papers are marked with patterns of colored dots in their corners for identification. A camera frame can be process to figure out where the dots are, what papers they correspond to, and therefore where papers are visible in the space. Projectors in the ceiling project graphics on the programs as the running programs instruct them to.

Paper sensing:
1. Grab a frame from the webcam and find everything that looks like a dot (#1600)
2. Get the list of dots and figure out what papers they map to using the knowledge that papers have only certains patterns of dots in the four corners of a paper (#1800)
3. Get the list of visual papers and wish that the corresponding program stored on a computer was running. (#826)
4. Get all wishes for programs that should be running and run them on a computer. (#1900)

Projection:
1. A process listens for the locations of all papers and each papers wishes for graphics to be drawn on them. Additionally it listens for a projector-camera calibration so the display is able to perform the projection mapping. (#1700)

To edit and create programs, there is a special piece of paper that edits the code of whatever paper it is pointing at. The code being edited is projected on to the text editor paper. A wireless keyboard associated with the text editor edits the text. When a new version of code is saved, a new version of the paper is printed out of a new sheet of paper to replace the old piece of paper.

Program editing:
1. The input from a wireless keyboard is captured and claimed to the room (#648)
2. When the system starts, a program reads the contents of all code files and claims them to the room (#390)
3. The text editor paper (#1013):
    a. Gets the source of the program it is editing
    b. Listens for the latest key presses to control the text editor cursor
    c. Wishes the text editor graphics would be projected on it
    d. When saving a program, the text editor wishes some other process would persist the new source code
4. A program listens for wishes of edited programs, transforms the code from the room's domain specific language into, and then wishes that the files on disk would be edited and run (#40). Another program persists the changes to the source code to the computer's disk (#577) and then causes the process to be restarted (#1900).
5. Simultaneously when saving a program, another program generates a PDF of a new piece of paper (#1382) and then another program talks to the printer to print the PDF (#498). 

### Alternative Inputs

Instead of sensing papers with cameras looking for colored dots, put a RFID card of the back of every paper and embedded RFID sensors in the room. A RFID sensor detecting a card is the equilalent of "the paper is in the room" and that is should be running.

### Alternative Outputs

Sounds, Stands of light, physical movement of things in the rooom, smells.
