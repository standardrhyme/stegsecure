import sys
import os 

# onlyLsb returns the LSB of x
def onlyLsb(x):
    return x & 1

# exceptLsb returns all the bits other than the last one of x
def exceptLsb(x):
    return x << 1

# updateParams updates the provided parameters, depending on the LSB and MSBs
def updateParams(u, v, params):
    uMsb = exceptLsb(u)
    uLsb = onlyLsb(u)
    vMsb = exceptLsb(v)
    vLsb = onlyLsb(v)

    # if only the LSB are different
    if (uMsb == vMsb) and (uLsb != vMsb):
        params[0] += 1

    # if they are the same
    if u == v:
        params[3] += 1

    if ((vLsb == 0) and (u < v)) or ((vLsb == 1) and (u > v)):
        params[1] += 1

    if ((vLsb == 0) and (u > v)) or ((vLsb == 1) and (u < v)):
        params[2] += 1

def analyzeSamplePairs(image):
    px = image.load()
    width, height = image.size

    # Based off of https://github.com/b3dk7/StegExpose/blob/master/SamplePairs.java
    avg = 0
    for color in range(3): # colors
        #         W  X  Y  Z
        params = [0, 0, 0, 0]
        P = 0

        # Compute horizontal pairs
        for y in range(height):
            for x in range(0, width - 1, 2):
                pair = [px[x, y], px[x + 1, y]]

                # Extract the color value for the pair pixels
                u = pair[0][color]
                v = pair[1][color]

                updateParams(u, v, params)

                P += 1

        # Compute vertical pairs
        for y in range(0, height - 1, 2):
            for x in range(width):
                pair = [px[x, y], px[x, y + 1]]

                # Extract the color value for the pair pixels
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

        # Solve for the largest root
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

        avg += x

    average = avg / 3
    probability = min(abs(average), 1)

    return probability > 0.5, probability

def main():
    from PIL import Image
    # Ask the user what image they would like to analyze
    print("FILE NAME: ", sys.argv[1])
    if os.path.isfile(sys.argv[1]) == False:
        print("The file does not exist.")
        sys.exit(1)
    imagename = sys.argv[1]



    #Open the image
    with Image.open(imagename) as im:
        result, probability = analyzeSamplePairs(im)

        print("Probability of being a stego image:", probability)
        if result:
            print("This is probably a stego image.")
        else:
            print("This is probably not a stego image.")

if __name__ == "__main__":
    main()
