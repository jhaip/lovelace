from helper2 import init, claim, retract, prehook, subscription, batch, MY_ID_STR, listen
import wx
import cv2
import time
import logging
import numpy as np
import math
import json
import zmq
import sys

CAM_WIDTH = 1920
CAM_HEIGHT = 1080

papers = []
projector_calibration = None
centered_labels = {}
texts = {}
lines = {}
draw_wishes = {}

def update_draw_wishes():
    global texts, lines, centered_labels, draw_wishes
    draw_wishes = {}
    # draw_wishes[source][target]
    for source in lines:
        if source not in draw_wishes:
            draw_wishes[source] = {}
            draw_wishes[source][source] = []
        draw_wishes[source][source] += lines[source]
    for source in texts:
        if source not in draw_wishes:
            draw_wishes[source] = {}
            draw_wishes[source][source] = []
        draw_wishes[source][source] += texts[source]
    for source in centered_labels:
        if source not in draw_wishes:
            draw_wishes[source] = {}
            draw_wishes[source][source] = []
        draw_wishes[source][source] += centered_labels[source]

def mapPaperResultToLegacyDataFormat(result):
    return {
        "id": result["id"],
        "corners": [
            {"CornerId": 0, "x": result["x1"], "y": result["y1"]},
            {"CornerId": 1, "x": result["x2"], "y": result["y2"]},
            {"CornerId": 2, "x": result["x3"], "y": result["y3"]},
            {"CornerId": 3, "x": result["x4"], "y": result["y4"]}
        ]
    }

def map_projector_calibration_to_legacy_data_format(result):
    return [
        [result["x1"], result["y1"]],
        [result["x2"], result["y1"]],
        [result["x3"], result["y1"]],
        [result["x4"], result["y1"]]
    ]

@subscription(["$ camera $cameraId sees paper $id at TL ($x1, $y1) TR ($x2, $y2) BR ($x3, $y3) BL ($x4, $y4) @ $time"])
def sub_callback_papers(results):
    global papers
    logging.info("sub_callback_papers")
    logging.info(results)
    papers = list(map(mapPaperResultToLegacyDataFormat, results))
    logging.info(papers)


@subscription(["$ camera $cameraId has projector calibration TL ($x1, $y1) TR ($x2, $y2) BR ($x3, $y3) BL ($x4, $y4) @ $time"])
def sub_callback_calibration(results):
    global projector_calibration
    logging.info("sub_callback_calibration")
    logging.info(results)
    projector_calibration = list(map(map_projector_calibration_to_legacy_data_format, results))
    logging.info(projector_calibration)

@subscription(["$id draw a ($r, $g, $b) line from ($x, $y) to ($xx, $yy)"])
def sub_callback_line(results):
    global lines
    logging.info("sub_callback_line")
    logging.info(results)
    lines = {}
    for v in results:
        source = int(v["id"])
        if source not in lines:
            lines[source] = []
        lines[source].append({"type": "line", "options": [v["x"], v["y"], v["xx"], v["yy"]]})
    logging.info(lines)
    update_draw_wishes()


@subscription(["$id draw $centered label $text at ($x, $y)"])
def sub_callback_centered_labels(results):
    global centered_labels
    logging.info("sub_callback_centered_labels")
    logging.info(results)
    centered_labels = {}
    for v in results:
        source = int(v["id"])
        if source not in centered_labels:
            centered_labels[source] = []
        centered_labels[source].append({
            "type": "text",
            "options": {
                "text": v["text"],
                "x": v["x"],
                "y": v["y"]
            }
        })
    logging.info(centered_labels)
    update_draw_wishes()


@subscription(["$id draw $size text $text at ($x, $y)"])
def sub_callback_text(results):
    global texts
    logging.info("sub_callback_text")
    logging.info(results)
    texts = {}
    for v in results:
        source = int(v["id"])
        if source not in texts:
            texts[source] = []
            texts[source].append({
                "type": "fontsize",
                "options": int(v["size"].replace('pt', ''))
            })
        texts[source].append({
            "type": "text",
            "options": {
                "text": v["text"],
                "x": v["x"],
                "y": v["y"]
            }
        })
    logging.info(texts)
    update_draw_wishes()

class Example(wx.Frame):
    ID_TIMER = 1
    GRAPHICS_TIMER = 2

    def __init__(self, parent, title):
        super(Example, self).__init__(parent, title=title,
                                      size=(CAM_WIDTH, CAM_HEIGHT))
        
        self.lastPaint = time.time()
        self.bigStart = time.time()

        self.i = 0
        self.bmp = None
        self.projection_matrix = None

        self.timer = wx.Timer(self, Example.ID_TIMER)
        self.Bind(wx.EVT_TIMER, self.OnTimer, id=Example.ID_TIMER)

        self.graphics_timer = wx.Timer(self, Example.GRAPHICS_TIMER)
        self.Bind(wx.EVT_TIMER, self.OnGraphicsTimer,
                  id=Example.GRAPHICS_TIMER)

        fps = 10
        self.timer.Start(1000./fps)

        graphics_fps = 10
        self.graphics_timer.Start(1000./graphics_fps)

        self.Bind(wx.EVT_PAINT, self.OnPaint)

        # self.Centre()
        self.Show()
        self.Maximize(True)

    def OnGraphicsTimer(self, event):
        self.Refresh()

    def OnTimer(self, event):
        global projector_calibration
        start = time.time()
        self.bigStart = time.time()
        # logging.error("loop")
        wishes = []
        deaths = []
        old_calibration = projector_calibration
        
        # logging.info("listening")
        listen()
        # logging.info("done listening")

        # self.onWish(wishes)
        # self.onProgramDeath(deaths)

        if old_calibration != projector_calibration:
            if projector_calibration and len(projector_calibration) is 4:
                logging.error("RECAL PROJECTION MATRIX")
                pts1 = np.float32(projector_calibration)
                pts2 = np.float32(
                    [[0, 0], [CAM_WIDTH, 0], [CAM_WIDTH, CAM_HEIGHT], [0, CAM_HEIGHT]])
                self.projection_matrix = cv2.getPerspectiveTransform(
                    pts1, pts2)

        end = time.time()
        # print(1000*(end - start), "ms", 1.0/(end - start), "fps")

    def OnPaint(self, e):
        global draw_wishes, papers, lines
        now = time.time()
        # if self.lastPaint:
        #     print(1.0/(now - self.lastPaint), "fps for paint")
        self.lastPaint = time.time()

        dc = wx.BufferedPaintDC(self)
        gc = wx.GraphicsContext.Create(dc)
        dc.SetBackground(wx.Brush(wx.Colour(0, 0, 0)))
        dc.Clear()
        dc.SetTextForeground(wx.Colour(255, 255, 255))
        dc.SetBrush(wx.Brush())
        font = dc.GetFont()
        font.SetWeight(wx.FONTWEIGHT_BOLD)
        dc.SetFont(font)

        # if self.bmp:
        #     dc.DrawBitmap(self.bmp, 0, 0)

        paper_draw_wishes = {}
        if draw_wishes:
            for wish_source in draw_wishes:
                for target in draw_wishes[wish_source]:
                    target_commands = draw_wishes[wish_source][target]
                    if target not in paper_draw_wishes:
                        paper_draw_wishes[target] = []
                    paper_draw_wishes[target].extend(target_commands)
        logging.error("PAPER DRAW WISHES:")
        logging.error(lines)
        logging.error(draw_wishes)
        logging.error("---draw wishes")
        logging.error(paper_draw_wishes)

        for i, target in enumerate(paper_draw_wishes):
            dc.DrawText(str(target) + ": " +
                        json.dumps(paper_draw_wishes[target]), 10, 10+16*i)

        self.draw_global_wishes(gc, paper_draw_wishes.get("global"))

        logging.error("PAPERS:")
        logging.error(papers)
        if papers:
            for paper in papers:
                if len(paper["corners"]) == 4:
                    self.draw_paper(
                        gc, paper, paper_draw_wishes.get(paper["id"]))

        end = time.time()
        # print(1000*(end - now), "ms", 1.0/(end - now), "fps to do paint stuff")

    def dist(self, p1, p2):
        return math.sqrt((p1["x"] - p2["x"])**2 + (p1["y"] - p2["y"])**2)

    def draw_global_wishes(self, gc, commands):
        if not commands:
            return
        self.draw_commands(gc, commands, CAM_WIDTH)

    def draw_commands(self, gc, draw_commands, width):
        paper_font = wx.Font(
            int(width/10), wx.FONTFAMILY_TELETYPE, wx.NORMAL, wx.BOLD)
        paper_font_color = wx.Colour(255, 255, 255)

        # img = wx.Image("./test_image.png", wx.BITMAP_TYPE_ANY)
        # bmp = gc.CreateBitmapFromImage(img)
        # gc.DrawBitmap(bmp, 100, 0, img.GetWidth(), img.GetHeight())
        gc.SetFont(paper_font, paper_font_color)
        # gc.DrawText("Paper "+str(paper["id"]), 0, 0)

        last_pen = wx.Pen("white")
        last_stroke_width = 1
        last_brush = wx.Brush("blue")
        gc.SetPen(last_pen)
        gc.SetBrush(last_brush)

        if draw_commands:
            print(draw_commands)
            for command in draw_commands:
                command_type = command.get("type")
                opt = command.get("options")
                if command_type == "rectangle":
                    if opt:
                        gc.DrawRectangle(
                            opt["x"], opt["y"], opt["w"], opt["h"])
                elif command_type == "ellipse":
                    if opt:
                        gc.DrawEllipse(opt["x"], opt["y"], opt["w"], opt["h"])
                elif command_type == 'text':
                    if opt:
                        print("DRAWING TEXT")
                        lines = opt["text"].split("\n")
                        line_height = paper_font.GetPixelSize().GetHeight() * 1.3
                        for i, l in enumerate(lines):
                            gc.DrawText(
                                l, opt["x"], opt["y"] + i * line_height)
                    else:
                        print("would draw text but missing opt")
                elif command_type == 'line':
                    if opt and len(opt) == 4:
                        # actually only drawing 1 line
                        gc.DrawLines([wx.Point2D(opt[0], opt[1]),
                                      wx.Point2D(opt[2], opt[3])])
                    else:
                        print("bad line")
                        print(opt)
                elif command_type == 'polygon':
                    if opt and len(opt) > 2:
                        path = gc.CreatePath()
                        path.MoveToPoint(wx.Point2D(opt[0][0], opt[0][1]))
                        for pt in opt[1:]:
                            path.AddLineToPoint(wx.Point2D(pt[0], pt[1]))
                        gc.DrawPath(path)
                    else:
                        print("bad polygon")
                        print(opt)
                elif command_type == 'fill':
                    if opt:
                        if type(opt) is str:
                            # color name like "blue"
                            last_brush = wx.Brush(opt)
                        elif len(opt) is 3:
                            last_brush = wx.Brush(
                                wx.Colour(opt[0], opt[1], opt[2]))  # RGB
                        else:
                            last_brush = wx.Brush(
                                wx.Colour(opt[0], opt[1], opt[2], opt[3]))  # RGBA
                        gc.SetBrush(last_brush)
                elif command_type == 'stroke':
                    if opt:
                        if type(opt) is str:
                            last_pen = wx.Pen(opt)  # color name like "blue"
                        elif len(opt) is 3:
                            last_pen = wx.Pen(
                                wx.Colour(opt[0], opt[1], opt[2]))  # RGB
                        else:
                            last_pen = wx.Pen(
                                wx.Colour(opt[0], opt[1], opt[2], opt[3]))  # RGBA
                        last_pen.SetWidth(last_stroke_width)
                        gc.SetPen(last_pen)
                elif command_type == 'nostroke':
                    last_pen.SetStyle(wx.PENSTYLE_TRANSPARENT)
                    gc.SetPen(last_pen)
                elif command_type == 'nofill':
                    last_brush.SetStyle(wx.BRUSHSTYLE_TRANSPARENT)
                    gc.SetBrush(last_brush)
                elif command_type == 'strokewidth':
                    if opt:
                        last_stroke_width = int(opt)
                        last_pen.SetWidth(last_stroke_width)
                        gc.SetPen(last_pen)
                elif command_type == 'fontsize':
                    if opt:
                        paper_font = wx.Font(
                            opt, wx.FONTFAMILY_TELETYPE, wx.NORMAL, wx.BOLD)
                        gc.SetFont(paper_font, paper_font_color)
                elif command_type == 'fontcolor':
                    if opt and len(opt) == 3:
                        paper_font_color = wx.Colour(opt[0], opt[1], opt[2])
                        gc.SetFont(paper_font, paper_font_color)
                elif command_type == 'push':
                    gc.PushState()
                elif command_type == 'pop':
                    gc.PushState()
                elif command_type == 'translate':
                    if opt:
                        gc.Translate(opt["x"], opt["y"])
                elif command_type == 'rotate':
                    if opt:
                        gc.Rotate(opt)
                elif command_type == 'scale':
                    if opt:
                        gc.Translate(opt["x"], opt["y"])
                else:
                    print("Unrecognized command:")
                    print(command)

    def draw_paper(self, gc, paper, draw_commands):
        if not draw_commands or len(draw_commands) == 0:
            return
        tl = None
        tr = None
        bl = None
        for corner in paper["corners"]:
            if corner["CornerId"] == 0:
                tl = self.project2(corner)
            elif corner["CornerId"] == 1:
                tr = self.project2(corner)
            elif corner["CornerId"] == 3:
                bl = self.project2(corner)
        paper_width = self.dist(tl, tr)
        paper_height = self.dist(tl, bl)
        paper_origin = tl
        logging.error("paper origin:")
        logging.error(paper_origin)
        paper_angle = math.atan2(tr["y"] - tl["y"], tr["x"] - tl["x"])

        gc.BeginLayer(1.0)

        gc.PushState()
        gc.Translate(paper_origin["x"], paper_origin["y"])
        gc.Rotate(paper_angle)

        gc.SetPen(wx.Pen("red", 3))
        # gc.SetBrush(wx.Brush("blue"))

        # gc.DrawRectangle(0, 0, paper_width, paper_height)

        self.draw_commands(gc, draw_commands, paper_width)

        gc.PopState()
        gc.EndLayer()

    def project2(self, _pt):
        pt = _pt.copy()
        # logging.error("1:")
        # logging.error(_pt)
        # logging.error("2:")
        # logging.error(pt)
        # return pt
        if self.projection_matrix is not None:
            # return pts
            pts = [(pt["x"], pt["y"])]
            dst = cv2.perspectiveTransform(
                np.array([np.float32(pts)]), self.projection_matrix)
            pt["x"] = int(dst[0][0][0])
            pt["y"] = int(dst[0][0][1])
            return pt
        # logging.error("MISSING PROJECTION MATRIX FOR PAPERS!")
        return pt


if __name__ == '__main__':
    init(__file__, skipListening=True)
    # while True:
    #     listen()
    #     time.sleep(1.0)
    app = wx.App()
    Example(None, 'Line')
    app.MainLoop()
