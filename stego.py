from PIL import Image
import binascii


#Ask the user for the name of the cover image that is in the current directory.
imagename = input("Image Name: ")

#Ask the user for the message they would like to hide within the input cover image
inputstring = input("Message: ")

#Append 'endmessage' to the user input. This will allow the 'unstego.py' to know when to stop scanning pixels.
inputstring = inputstring + "endmessage"

#Turn the message from string to condensed binary (0101010101010101)
message = ''.join(format(ord(i), '08b') for i in inputstring)

#Get the size of the input binary
lenmessage = len(message)


# - - - - - - IMAGE - - - - - - - #
#Open the image
with Image.open(imagename) as im:
    px = im.load()    

#Get the size of the image (h x w)
imagesize = im.size
imagewidth = imagesize[0]
imageheight = imagesize[1]

#Check how many pixels would need to be altered to fit the message
if lenmessage%3==0:
    pixelsneeded = lenmessage // 3
else:
    pixelsneeded = (lenmessage // 3)+1

#If the number of pixels needed is less than the number of pixels of the image, the message will fit.
if pixelsneeded <= imageheight*imagewidth:
    #Keep track of how many bits have been read into the image pixels
    messagecount = 0
    #Loop through pixels
    for row in range(imageheight):
        if messagecount>=lenmessage: break
        else:
            for col in range(imagewidth):
                if messagecount>=lenmessage: break
                else:
                    #Loop through the rgb values of each pixel
                    for rgb in range(0,3):
                        if messagecount>=lenmessage: break
                        else:
                            #Get the value of the color for the specific pixel and color channel
                            old = px[row,col][rgb]
                                
                            #Is the corresponding message character a 0 or 1?
                            if message[messagecount]=="1":
                                #If the message has a value of 1 yet the color value is even (meaning the bits would end in 0), add one to the color value
                                if px[row,col][rgb]%2==0:
                                    #Change color channel values tuple into list
                                    listcolorvalues=list(px[row,col])
                                    
                                    #Add one to color value
                                    listcolorvalues[rgb]+=1

                                    #Change back to tuple and set pixel color equal to the new values
                                    px[row,col]=tuple(listcolorvalues)
                            elif message[messagecount]=="0":
                                #If the message has a value of 0 yet the color value is odd (meaning the bits would end in 1), subtract one from the color value
                                if px[row,col][rgb]%2==1:
                                    #Change color channel values tuple into list
                                    listcolorvalues=list(px[row,col])
                                    
                                    #Subtract one from color value
                                    listcolorvalues[rgb]-=1

                                    #Change back to tuple and set pixel color equal to the new values
                                    px[row,col]=tuple(listcolorvalues)
                            #Move forward in the message
                            messagecount+=1

    print("\n- - - The image pixels have been altered to conceal the input message. The image will now save. - - -")
    #Save the image
    im.save("samplestego.png", format="png")
    #Let the user know the image has been saved
    print("- - - The rendered image has been saved as 'samplestego.png' in the current directory. Thank you. - - - ")       
#If the message does not fit, let the user know.
else:
    print("- - - The message would need to alter:", pixelsneeded,"pixels. ~ Therefore, the message will not fit in the currently selected image. Try again. - - -")


                    
                
                
            
    











