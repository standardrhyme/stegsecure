from PIL import Image
import binascii

#Ask the user for the name of the image they would like to use
imagename = input("Image Name: ")

#Ask the user for the text input
inputstring = input("Message: ")

#Open the image
with Image.open(imagename) as im:
    px = im.load()

#Transform the input from string to binary
inputbinary = ' '.join(format(ord(i), '08b') for i in inputstring)
inputstring = "".join([chr(int(binary, 2)) for binary in inputbinary.split(" ")])
print("Your message was:", inputstring)
                       
message = ''.join(format(ord(i), '08b') for i in inputstring)

#Get the size of the input binary
lenmessage = len(message)

#Get the size of the image (h x w)
imagesize = im.size
imageheight = imagesize[0]
imagewidth = imagesize[1]

#Check how many pixel changes are needed
if lenmessage%3==0:
    pixelsneeded = lenmessage // 3
else:
    pixelsneeded = (lenmessage // 3)+1

#Check if the message can fit in the image
if pixelsneeded <= imageheight*imagewidth:
    messagecount = 0
    #loop through pixels
    for i in range(imageheight):
        if messagecount>=lenmessage: break
        else:
            for j in range(imagewidth):
                if messagecount>=lenmessage: break
                else:
                    for k in range(0,3):
                        if messagecount>=lenmessage: break
                        else:
                            #this prints pixel by pixel, r then g then b
                            old = px[i,j][k]
                                
                            #should the pixel color end in 0 or 1 (according to message)
                            if message[messagecount]=="1":
                                #does the current pixel color end in 1? if px[i,j][k]%2==1: do nothing
                                if px[i,j][k]%2==0:
                                    #change tuple into list
                                    listcolorvalues=list(px[i,j])
                                    
                                    #if it ends in 0, add 1 to color
                                    listcolorvalues[k]+=1

                                    #change to tuple
                                    px[i,j]=tuple(listcolorvalues)
                            elif message[messagecount]=="0":
                                #does the current pixel color end in 0? if px[i,j][k]%2==0: do nothing
                                if px[i,j][k]%2==1:
                                    #change tuple into list
                                    listcolorvalues=list(px[i,j])
                                    
                                    #if it ends in 0, subtract 1 to color
                                    listcolorvalues[k]-=1

                                    #change to tuple
                                    px[i,j]=tuple(listcolorvalues)
                            messagecount+=1

    im.save("samplestego.png", format="png")
    print("samplestego.png has been saved to the current directory. Thank you. \n\n")       
else:
    print("The message would need to alter:", pixelsneeded,"pixels. ~ Therefore, the message will not fit in the currently selected image. Try again.")


                    
                
                
            
    











