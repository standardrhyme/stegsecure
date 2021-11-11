from PIL import Image
import binascii

def onlyLsb(x):
    return x & 0x1

def exceptLsb(x):
    return x << 1

def updateParams(u, v, params):
    uMsb = exceptLsb(u)
    uLsb = onlyLsb(u)
    vMsb = exceptLsb(v)
    vLsb = onlyLsb(v)

    if (uMsb == vMsb) and (uLsb != vMsb):
        params[0] += 1

    if u == v:
        params[3] += 1

    if ((vLsb == 0) and (u < v)) or ((vLsb == 1) and (u > v)):
        params[1] += 1

    if ((vLsb == 0) and (u > v)) or ((vLsb == 1) and (u < v)):
        params[2] += 1

# Ask the user what image they would like to analyze
imagename = input("Enter the name of the image you would like to analyze: ")

#Open the image
with Image.open(imagename) as im:
    px = im.load()
    height, width = im.size

# Based off of https://github.com/b3dk7/StegExpose/blob/master/SamplePairs.java
avg = 0
for color in range(3): # colors
    #         W  X  Y  Z
    params = [0, 0, 0, 0]
    P = 0
    for y in range(height):
        for x in range(0, width - 1, 2):
            pair = [px[y,x], px[y, x + 1]]

            u = pair[0][color]
            v = pair[1][color]

            updateParams(u, v, params)

            P += 1

    for y in range(0, height - 1, 2):
        for x in range(width):
            pair = [px[y,x], px[y + 1, x]]

            u = pair[0][color]
            v = pair[1][color]

            updateParams(u, v, params)

            P += 1

    W, X, Y, Z = params

    a = (W + Z) / 2
    b = (2 * X) - P
    c = Y - X

    if a == 0:
        x = c / b

    discriminant = b**2 - (4*a*c)
    if discriminant >= 0:
        posroot = ((-1*b) + discriminant**0.5) / (2*a)
        negroot = ((-1*b) - discriminant**0.5) / (2*a)

        if abs(posroot) <= abs(negroot):
            x = posroot
        else:
            x = negroot
    else:
        x = c / b

    print("Color", color, "result:", x)

    avg += x

final = avg / 3
print("Probability of being a stego image:", final)
