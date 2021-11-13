from PIL import Image
import binascii

#Ask the user what image they would like to unstego
imagename = input("Enter the name of the image you would like to unstego: ")

#Open the image
with Image.open(imagename) as im:
    px = im.load()

#Get the size of the image (w x h)
imagesize = im.size
imagewidth = imagesize[0]
imageheight = imagesize[1]

message=""
nextcharacter=""
count = 0
goon = True
pixelsadjusted=1
#loop through pixels
for i in range(imageheight):
    if goon:
        for j in range(imagewidth):
            if goon:
                for k in range(3):
                    if goon:
                        if count==8:
                            pixelsadjusted = int(nextcharacter, 2)
                            nextcharacter=''
                        #if the message has been complete uncoded
                        elif count>=((pixelsadjusted)*3)+8 and count%8==0:
                            goon = False
                        elif count>8 and count%8==0:
                            nextcharacterbinary = int(nextcharacter,2) #get the integer value of that character
                            charactertoadd = chr(nextcharacterbinary) #get the character value
                            #add to guess
                            message+=charactertoadd
                            #reset count, nonchar, and next char
                            nextcharacter=""
                        #if not enough pixels, add one to count, get the next value
                        count+=1
                        old = px[j,i][k]
                        value = str(old%2)
                        nextcharacter+=value
                    else: break
            else: break
    else: break       

print("\n- - - The concealed message was: {", message, "} - - -")


            
                    
                
                
            
    











