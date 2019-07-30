from imutils.video import WebcamVideoStream
import numpy as np
import cv2
import imutils

capture = WebcamVideoStream(src=0)
capture.start()
image = capture.read()
# image = imutils.resize(image, width=400)

pts = np.array([(80*4.8, 4*4.8), (305*4.8, 1*4.8), (344*4.8, 224*4.8), (46*4.8, 225*4.8)], dtype = "float32")

maxWidth = 1430
maxHeight = 1086
dst = np.array([
		[0, 0],
		[maxWidth - 1, 0],
		[maxWidth - 1, maxHeight - 1],
		[0, maxHeight - 1]], dtype = "float32")
M = cv2.getPerspectiveTransform(pts, dst)
warped = cv2.warpPerspective(image, M, (maxWidth, maxHeight))

warped = cv2.flip( warped, -1 ) # flip both axes

warped_grey = cv2.cvtColor(warped, cv2.COLOR_BGR2GRAY)
# ret,thresh1 = cv2.threshold(warped_grey, 150, 255, cv2.THRESH_BINARY)
# th2 = cv2.adaptiveThreshold(warped_grey, 255, cv2.ADAPTIVE_THRESH_GAUSSIAN_C, cv2.THRESH_BINARY, 115, 1)
# ret2,th3 = cv2.threshold(warped_grey,0,255,cv2.THRESH_BINARY+cv2.THRESH_OTSU)

# th4 = cv2.adaptiveThreshold(warped_grey,255,cv2.ADAPTIVE_THRESH_MEAN_C,\
#             cv2.THRESH_BINARY,11,2)

cv2.imshow("Original", image)
cv2.imshow("Warped", warped)
# cv2.imshow("Threshold1", thresh1)
# cv2.imshow("Threshold2", th2)
# cv2.imshow("Threshold3", th3)
# cv2.imshow("Threshold4", th4)

GRID_WIDTH_CELLS = 6
GRID_HEIGHT_CELLS = 4
CELL_WIDTH_PX = 155
CELL_HEIGHT_PX = 135
ORIGIN_X = 25
ORIGIN_Y = 10
CELL_X_PADDING_PX = 90
CELL_Y_PADDING_PX = 150

for ix in range(GRID_WIDTH_CELLS):
    for iy in range(GRID_HEIGHT_CELLS):
        x = ORIGIN_X + ix * (CELL_WIDTH_PX + CELL_X_PADDING_PX)
        y = ORIGIN_Y + iy * (CELL_HEIGHT_PX + CELL_Y_PADDING_PX)
        x2 = x + CELL_WIDTH_PX
        y2 = y + CELL_HEIGHT_PX
        cv2.rectangle(warped,(x,y),(x2,y2),(0,255,0),3)

cv2.imshow("Warped2", warped)
# w = 36*5
# h = 31*5
# x = 5*5
# y = 20*5

# roi = warped_grey[y:(y+h), x:(x+w)]
# cv2.imshow("ROI", roi)

cv2.waitKey(0)