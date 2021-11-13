from PIL import Image
import binascii

def stegoImage(image, inputstring, output):
    px = image.load()
    #Get the size of the image (h x w)
    imagesize = image.size
    imagewidth = imagesize[0]
    imageheight = imagesize[1]
    

    #Turn the message from string to condensed binary (0101010101010101)
    message = ''.join(format(ord(i), '08b') for i in inputstring)

    #Get the size of the input binary
    lenmessage = len(message)

    #Check how many pixels would need to be altered to fit the message
    if lenmessage % 3 == 0:
        pixelsneeded = lenmessage // 3
    else:
        pixelsneeded = (lenmessage // 3)+1
    
    binpixelsneeded = format(pixelsneeded,'08b')
    message = binpixelsneeded+message

    #If the number of pixels needed is less than the number of pixels of the image, the message will fit.
    if pixelsneeded > imageheight*imagewidth:
        print("- - - The message would need to alter:", pixelsneeded,"pixels. ~ Therefore, the message will not fit in the currently selected image. Try again. - - -")
        return False

    #Keep track of how many bits have been read into the image pixels
    messagecount = 0
    #Loop through pixels
    for i in range(imageheight):
        for j in range(imagewidth):
            #Loop through the rgb values of each pixel
            for rgb in range(0,3):
                if messagecount-8>=lenmessage:
                    break
                listcolorvalues=list(px[j,i])
                # Set the LSB to 0
                listcolorvalues[rgb] = (listcolorvalues[rgb] // 2) * 2
                # Change the LSB to 1 if necessary
                if message[messagecount] == "1":
                    listcolorvalues[rgb] += 1
                #Change back to tuple and set pixel color equal to the new values
                px[j,i]=tuple(listcolorvalues)
                #Move forward in the message
                messagecount+=1

    print("\n- - - The image pixels have been altered to conceal the input message. The image will now save. - - -")
    #Save the image
    image.save(output, format="png")
    #Let the user know the image has been saved
    print("- - - The rendered image has been saved as 'samplestego.png' in the current directory. Thank you. - - - ")

    return True

def main():
    #Ask the user for the name of the cover image that is in the current directory.
    imagename = input("Cover image name: ")

    #Ask the user if they would like to give a .txt file or manually enter a message.
    # inputchoice = input("Input choice: txt file (1) or manual input (2)?: ")
    inputchoice = '2'

    if inputchoice == '1':
        #Ask the user to input the txt file name
        textfilename = input("Txt File Name: ")
        #Open the txt file
        with open(textfilename, 'r') as file:
            data = file.read().replace('\n', '')
            inputstring = data


    elif inputchoice == '2':
        #Ask the user for the message they would like to hide within the input cover image
        inputstring = input("Message: ")


    # - - - - - - IMAGE - - - - - - - #
    #Open the image
    with Image.open(imagename) as im:
        stegoImage(im, inputstring, "samplestego.png")

if __name__ == "__main__":
    main()
