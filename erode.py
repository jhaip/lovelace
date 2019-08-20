import cv2
import numpy as np

images = [
    cv2.imread('/Users/jhaip/Code/lovelace/src/standalone_processes/files/cv_tiles/up.png',0),
    cv2.imread('/Users/jhaip/Code/lovelace/src/standalone_processes/files/cv_tiles/down.png',0),
    cv2.imread('/Users/jhaip/Code/lovelace/src/standalone_processes/files/cv_tiles/right.png',0),
    cv2.imread('/Users/jhaip/Code/lovelace/src/standalone_processes/files/cv_tiles/left.png',0),
    cv2.imread('/Users/jhaip/Code/lovelace/src/standalone_processes/files/cv_tiles/loopstart.png',0),
    cv2.imread('/Users/jhaip/Code/lovelace/src/standalone_processes/files/cv_tiles/loopstop.png',0),
]

def draw_style1():
    modimages = cv2.hconcat(images)
    cv2.imshow('style1', modimages)

def draw_style2():
    kernel = np.ones((5,5),np.uint8)
    modimages = []
    for m in images:
        t = cv2.erode(m,kernel,iterations=4)
        modimages.append(t)
    combined = cv2.hconcat(modimages)
    cv2.imshow('style2', combined)

def draw_style3():
    kernel = np.ones((3,3),np.uint8)
    modimages = []
    for m in images:
        t = cv2.erode(m,kernel,iterations=6)
        modimages.append(t)
    combined = cv2.hconcat(modimages)
    cv2.imshow('style3', combined)

def draw_style4():
    modimages = []
    for m in images:
        cdst = cv2.cvtColor(m, cv2.COLOR_GRAY2BGR)
        edges = cv2.Canny(m,50,150,apertureSize = 3)
        lines = cv2.HoughLines(edges,1,np.pi/180,25)
        if lines is not None:
            for line in lines:
                for rho,theta in line:
                    a = np.cos(theta)
                    b = np.sin(theta)
                    x0 = a*rho
                    y0 = b*rho
                    x1 = int(x0 + 1000*(-b))
                    y1 = int(y0 + 1000*(a))
                    x2 = int(x0 - 1000*(-b))
                    y2 = int(y0 - 1000*(a))

                    cv2.line(cdst,(x1,y1),(x2,y2),(0,0,255),2)
        modimages.append(cdst)
    combined = cv2.hconcat(modimages)
    cv2.imshow('lines', combined)

def draw_style5():
    modimages = []
    for m in images:
        size = np.size(m)
        skel = np.zeros(m.shape,np.uint8)
        element = cv2.getStructuringElement(cv2.MORPH_CROSS,(3,3))
        done = False
        while(not done):
            eroded = cv2.erode(m,element)
            temp = cv2.dilate(eroded,element)
            temp = cv2.subtract(m,temp)
            skel = cv2.bitwise_or(skel,temp)
            m = eroded.copy()

            zeros = size - cv2.countNonZero(m)
            if zeros==size:
                done = True
        modimages.append(skel)
    combined = cv2.hconcat(modimages)
    cv2.imshow('skel', combined)

def draw_contours():
    kernel = np.ones((7,7),np.uint8)
    modimages = []
    modimages2 = []
    for m in images:
        cdst = cv2.cvtColor(m, cv2.COLOR_GRAY2BGR)
        dilation = cv2.dilate(m,kernel,iterations = 1)
        image, contours, hierarchy = cv2.findContours(dilation,cv2.RETR_TREE,cv2.CHAIN_APPROX_SIMPLE)    
        biggest_contours = sorted(contours, key = cv2.contourArea, reverse = True)[:1] # get largest contour
        # cdst = cv2.drawContours(cdst, contours, -1, (0,255,0), 3)
        cdst = cv2.drawContours(cdst, biggest_contours, -1, (0,255,0), 3)
        modimages.append(cdst)
        c = biggest_contours[0]
        # determine the most extreme points along the contour
        extLeft = tuple(c[c[:, :, 0].argmin()][0])
        extRight = tuple(c[c[:, :, 0].argmax()][0])
        extTop = tuple(c[c[:, :, 1].argmin()][0])
        extBot = tuple(c[c[:, :, 1].argmax()][0])
        cv2.rectangle(cdst,(extLeft[0],extTop[1]),(extRight[0],extBot[1]),(0,255,0),1)

        closing = cv2.morphologyEx(m, cv2.MORPH_CLOSE, kernel)
        erosion = cv2.erode(closing,kernel,iterations = 1)
        dilation_copy = m.copy()
        cropped = dilation_copy[extTop[1]:extBot[1], extLeft[0]:extRight[0]]
        dim = (40, 40)
        resized = cv2.resize(cropped, dim, interpolation = cv2.INTER_NEAREST)
        modimages2.append(resized)

    combined = cv2.hconcat(modimages)
    cv2.imshow('contours', combined)
    combined2 = cv2.hconcat(modimages2)
    cv2.imshow('contours_cropped', combined2)


# cv2.imshow('image',img)
# cv2.imshow('image1', erosion1)
# cv2.imshow('image2',erosion10)
# cv2.imshow('image3', erosion3)
# cv2.imshow('erosion11', erosion11)
draw_style1()
draw_style2()
draw_style3()
draw_style4()
draw_style5()
draw_contours()

cv2.waitKey(0)
cv2.destroyAllWindows()
