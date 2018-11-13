# Example from https://stackoverflow.com/questions/14804741/opencv-integration-with-wxpython
import wx
from imutils.video import WebcamVideoStream
import imutils
import cv2
import time
import json
# import RPCClient
import zmq
import logging
import sys
import os

MY_ID_STR = None
CAM_WIDTH = 1920
CAM_HEIGHT = 1080

def initLogToFile(root_filename):
    global MY_ID_STR
    scriptName = os.path.basename(root_filename)
    scriptNameNoExtension = os.path.splitext(scriptName)[0]
    fileDir = os.path.dirname(os.path.realpath(root_filename))
    logPath = os.path.join(fileDir, 'logs/' + scriptNameNoExtension + '.log')
    logging.basicConfig(filename=logPath, level=logging.INFO)
    MY_ID = (scriptName.split(".")[0]).split("__")[0]
    MY_ID_STR = str(MY_ID).zfill(4)
    print("INSIDE INIT:")
    print(MY_ID)
    print(MY_ID_STR)
    print(logPath)

class ShowCapture(wx.Panel):
    def __init__(self, parent, capture, fps=1):
        wx.Panel.__init__(self, parent)

        initLogToFile(__file__)
        logging.error("begin")

        self.capture = capture
        ret, frame = (True, self.capture.read())

        height, width = frame.shape[:2]
        parent.SetSize((width, height))
        frame = cv2.cvtColor(frame, cv2.COLOR_BGR2RGB)

        self.bmp = wx.Bitmap.FromBuffer(width, height, frame)
        self.dots = []

        self.projector_calibration_state = None
        self.blob_detector = self.createSimpleBlobDetector()

        # self.M = RPCClient.RPCClient()
        # self.M.set_pub_high_water_mark(0)

        self.projector_calibration = [(50, 50), (CAM_WIDTH-50, 50),
                                      (CAM_WIDTH-50, CAM_HEIGHT-50),
                                      (50, CAM_HEIGHT-50)]
        
        logging.error("PRE ZMQ SETUP")

        time.sleep(1.0)
        context = zmq.Context()
        rpc_url = "localhost"
        self.pub_socket = context.socket(zmq.PUB)
        self.pub_socket.connect("tcp://{0}:5556".format(rpc_url))
        time.sleep(1.0)

        logging.error("post connect")

        self.claimProjectorCalibration()
        
        logging.error("post claim")

        self.timer = wx.Timer(self)
        self.timer.Start(1000./fps)

        self.Bind(wx.EVT_PAINT, self.OnPaint)
        self.Bind(wx.EVT_TIMER, self.NextFrame)

        self.Bind(wx.EVT_KEY_DOWN, self.OnKeyDown)
        self.Bind(wx.EVT_KEY_UP, self.OnKeyDown)
        self.Bind(wx.EVT_CHAR, self.OnKeyDown)
        self.Bind(wx.EVT_LEFT_UP, self.onClick)
        self.SetFocus()
    
    def claim(self, fact_string):
        global MY_ID_STR
        self.pub_socket.send_string("....CLAIM{}{}".format(
            MY_ID_STR, fact_string), zmq.NOBLOCK)
    
    def claimProjectorCalibration(self):
        global MY_ID_STR
        batch_claims = [{"type": "claim", "fact": [
            ["id", MY_ID_STR],
            ["text", "camera"],
            ["integer", "1"],
            ["text", "has"],
            ["text", "projector"],
            ["text", "calibration"],
            ["text", "TL"],
            ["text", "("],
            ["integer", str(self.projector_calibration[0][0])],
            ["text", ","],
            ["integer", str(self.projector_calibration[0][1])],
            ["text", ")"],
            ["text", "TR"],
            ["text", "("],
            ["integer", str(self.projector_calibration[1][0])],
            ["text", ","],
            ["integer", str(self.projector_calibration[1][1])],
            ["text", ")"],
            ["text", "BR"],
            ["text", "("],
            ["integer", str(self.projector_calibration[2][0])],
            ["text", ","],
            ["integer", str(self.projector_calibration[2][1])],
            ["text", ")"],
            ["text", "BL"],
            ["text", "("],
            ["integer", str(self.projector_calibration[3][0])],
            ["text", ","],
            ["integer", str(self.projector_calibration[3][1])],
            ["text", ")"],
            ["text", "@"],
            ["integer", str(int(round(time.time() * 1000)))],
        ]}]
        self.batch(batch_claims)

    def retract(self, fact_string):
        global MY_ID_STR
        print(fact_string)
        self.pub_socket.send_string("..RETRACT{}{}".format(
            MY_ID_STR, fact_string), zmq.NOBLOCK)
    
    def batch(self, batch_claims):
        global MY_ID_STR
        self.pub_socket.send_string("....BATCH{}{}".format(
            MY_ID_STR, json.dumps(batch_claims)), zmq.NOBLOCK)

    def OnPaint(self, evt):
        dc = wx.BufferedPaintDC(self)
        dc.SetBrush(wx.Brush())
        font =  dc.GetFont()
        font.SetWeight(wx.FONTWEIGHT_BOLD)
        dc.SetFont(font)

        dc.DrawBitmap(self.bmp, 0, 0)

        for dot in self.dots:
            dc.SetBrush(wx.Brush(wx.Colour(dot["color"][0], dot["color"][1], dot["color"][2])))
            # dc.SetBrush(wx.Brush(wx.Colour(255, 0, 0)))
            dc.SetPen(wx.Pen(wx.Colour(255, 0, 0)))
            s = 3
            dc.DrawEllipse(int(dot["x"])-s, int(dot["y"])-s, s*2, s*2)

        dc.SetBrush(wx.Brush(wx.Colour(0,255,255), style=wx.BRUSHSTYLE_TRANSPARENT))
        dc.SetPen(wx.Pen(wx.Colour(0,0,255)))
        dc.DrawPolygon(self.projector_calibration)

        if self.projector_calibration_state is not None:
            pt = self.projector_calibration[self.projector_calibration_state]
            dc.SetBrush(wx.Brush(wx.Colour(0,255,255), style=wx.BRUSHSTYLE_TRANSPARENT))
            dc.SetPen(wx.Pen(wx.Colour(255, 255, 255)))
            s = 3
            dc.DrawEllipse(int(pt[0])-s, int(pt[1])-s, s*2, s*2)
            dc.SetPen(wx.Pen(wx.Colour(0, 0, 0)))
            s = s + 1
            dc.DrawEllipse(int(pt[0])-s, int(pt[1])-s, s*2, s*2)
            dc.SetPen(wx.Pen(wx.Colour(255, 255, 255)))
            s = s + 1
            dc.DrawEllipse(int(pt[0])-s, int(pt[1])-s, s*2, s*2)
            dc.DrawText("EDITING CORNER " + str(self.projector_calibration_state), CAM_WIDTH/2, CAM_HEIGHT/2)

    def createSimpleBlobDetector(self):
        params = cv2.SimpleBlobDetector_Params()
        params.minThreshold = 50 # 150
        params.maxThreshold = 230 # 200
        params.filterByCircularity = True
        params.minCircularity = 0.5
        params.filterByArea = True
        params.minArea = 9
        params.filterByInertia = False
        is_v2 = cv2.__version__.startswith("2.")
        if is_v2:
            detector = cv2.SimpleBlobDetector(params)
        else:
            detector = cv2.SimpleBlobDetector_create(params)
        return detector

    def NextFrame(self, event):
        global MY_ID_STR
        start = time.time()

        ret, frame = (True, self.capture.read())
        if ret:
            keypoints = self.blob_detector.detect(frame)

            # print(self.keypoints)
            def keypointMapFunc(keypoint):
                # color = frame[int(keypoint.pt[1]), int(keypoint.pt[0])]
                colorSum = [0, 0, 0]
                N_H_SAMPLES = 1
                N_V_SAMPLES = 1
                TOTAL_SAMPLES = (2*N_H_SAMPLES+1) * (2*N_V_SAMPLES+1)
                for i in range(-N_H_SAMPLES, N_H_SAMPLES+1):
                    for j in range(-N_V_SAMPLES, N_V_SAMPLES+1):
                        color = frame[int(keypoint.pt[1])+i, int(keypoint.pt[0])+j]
                        colorSum[0] += int(color[0])
                        colorSum[1] += int(color[1])
                        colorSum[2] += int(color[2])
                return {
                    "x": int(keypoint.pt[0]),
                    "y": int(keypoint.pt[1]),
                    "color": [int(colorSum[2]/TOTAL_SAMPLES),
                              int(colorSum[1]/TOTAL_SAMPLES),
                              int(colorSum[0]/TOTAL_SAMPLES)]
                }
            self.dots = list(map(keypointMapFunc, keypoints))
            # print(self.dots)
            # self.M.claim("global", "dots", self.dots)

            
            # self.retract("$z dots $x $y color $a $b $c $t")
            # for dot in self.dots:
            #     self.claim("dots {} {} color {} {} {} {}".format(
            #         dot["x"], dot["y"], dot["color"][0], dot["color"][1], dot["color"][2], int(time.time()*1000.0)))
            batch_claims = [{"type": "retract", "fact": [
                            ["id", MY_ID_STR],
                            ["postfix", ""]
                            ]}]
            for dot in self.dots:
                batch_claims.append({"type": "claim", "fact": [
                    ["id", MY_ID_STR], ["text", "dots"],
                    ["float", str(dot["x"])], ["float", str(dot["y"])],
                    ["text", "color"],
                    ["float", str(dot["color"][0])], ["float", str(dot["color"][1])], ["float", str(dot["color"][1])],
                    ["integer", str(int(time.time()*1000.0))]
                ]})
            self.batch(batch_claims)

            frame = cv2.cvtColor(frame, cv2.COLOR_BGR2RGB)
            self.bmp.CopyFromBuffer(frame)
            #
            # img = wx.Bitmap.ConvertToImage( self.bmp )
            # img_str = img.GetData()
            # self.M.set_image(img_str)
            #
            self.Refresh()

        end = time.time()
        logging.error("{} {} fps".format(end - start, 1.0/(end - start)))

    def moveCurrentCalibrationPointRel(self, dx, dy):
        if self.projector_calibration_state is not None:
            prev = self.projector_calibration[self.projector_calibration_state]
            next = (prev[0] + dx, prev[1] + dy)
            self.projector_calibration[self.projector_calibration_state] = next
            self.claimProjectorCalibration()

    def moveCurrentCalibrationPoint(self, pt):
        if self.projector_calibration_state is not None:
            self.projector_calibration[self.projector_calibration_state] = (pt[0], pt[1])
            self.claimProjectorCalibration()

    def changeCurrentCalibrationPoint(self, key):
        if key == '`':
            self.projector_calibration_state = None
        elif key in ['1', '2', '3', '4']:
            self.projector_calibration_state = int(key)-1

    def OnKeyDown(self, event=None):
        keyCode = event.GetKeyCode()
        if keyCode == wx.WXK_UP:
            self.moveCurrentCalibrationPointRel(0, -1)
        elif keyCode == wx.WXK_RIGHT:
            self.moveCurrentCalibrationPointRel(1, 0)
        elif keyCode == wx.WXK_DOWN:
            self.moveCurrentCalibrationPointRel(0, 1)
        elif keyCode == wx.WXK_LEFT:
            self.moveCurrentCalibrationPointRel(-1, 0)
        else:
            unicodeKey = chr(event.GetUnicodeKey())
            if unicodeKey in ['`', '1', '2', '3', '4']:
                self.changeCurrentCalibrationPoint(unicodeKey)

    def onClick(self, event=None):
        if event:
            pt = event.GetPosition()
            self.moveCurrentCalibrationPoint(pt)


# capture = cv2.VideoCapture(0)
# capture.set(cv2.CAP_PROP_FRAME_WIDTH, CAM_WIDTH)
# capture.set(cv2.CAP_PROP_FRAME_HEIGHT, CAM_HEIGHT)
capture = WebcamVideoStream(src=0)
# time.sleep(2)
logging.error("Default settings:")
logging.error(capture.stream.get(cv2.CAP_PROP_FRAME_WIDTH))
logging.error(capture.stream.get(cv2.CAP_PROP_FRAME_HEIGHT))
logging.error(capture.stream.get(cv2.CAP_PROP_FOURCC))
logging.error(capture.stream.get(cv2.CAP_PROP_FORMAT))
logging.error(capture.stream.get(cv2.CAP_PROP_MODE))
logging.error(capture.stream.get(cv2.CAP_PROP_SETTINGS))
logging.error(capture.stream.get(cv2.CAP_PROP_FPS))
logging.error(capture.stream.get(cv2.CAP_PROP_BRIGHTNESS))
logging.error(capture.stream.get(cv2.CAP_PROP_CONTRAST))
logging.error(capture.stream.get(cv2.CAP_PROP_SATURATION))
logging.error(capture.stream.get(cv2.CAP_PROP_HUE))
logging.error(capture.stream.get(cv2.CAP_PROP_GAIN))
logging.error(capture.stream.get(cv2.CAP_PROP_EXPOSURE))
logging.error(capture.stream.get(cv2.CAP_PROP_AUTOFOCUS))
logging.error("---")

# capture.start()
# capture.read()

# time.sleep(2)
capture.stream.set(cv2.CAP_PROP_FRAME_WIDTH, CAM_WIDTH)
capture.stream.set(cv2.CAP_PROP_FRAME_HEIGHT, CAM_HEIGHT)
capture.stream.set(cv2.CAP_PROP_FOURCC, cv2.VideoWriter_fourcc('M', 'P', 'E', 'G'))
# capture.stream.set(cv2.CAP_PROP_FPS, TODO)
# capture.stream.set(cv2.CAP_PROP_BRIGHTNESS, 0)
# capture.stream.set(cv2.CAP_PROP_CONTRAST, TODO)
# capture.stream.set(cv2.CAP_PROP_SATURATION, TODO)
# capture.stream.set(cv2.CAP_PROP_HUE, TODO)
# capture.stream.set(cv2.CAP_PROP_GAIN, TODO)
# capture.stream.set(cv2.CAP_PROP_EXPOSURE, TODO)
time.sleep(2)
capture.start()
time.sleep(2)

# for i in range(10):
#     capture.read()

# initLogToFile(__file__)

app = wx.App()
frame = wx.Frame(None)
cap = ShowCapture(frame, capture)
frame.Show()
app.MainLoop()
