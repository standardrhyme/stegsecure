from PIL import Image
import binascii

#Ask the user what image they would like to unstego
imagename = input("Enter the name of the image you would like to unstego: ")

#Open the image
with Image.open(imagename) as im:
    px = im.load()

#Get the size of the image (h x w)
imagesize = im.size
imagewidth = imagesize[0]
imageheight = imagesize[1]

message=""
nextcharacter=""
count = 1
goon = True
#loop through pixels
for i in range(imagewidth):
    if goon:
        for j in range(imageheight):
            if goon:
                for k in range(3):
                    if count==9: #if we have enough pixel information to produce a character
                        nextcharacterbinary = int(nextcharacter,2) #get the integer value of that character
                        charactertoadd = chr(nextcharacterbinary) #get the character value
                        #add to guess
                        message+=charactertoadd
                        #check to make sure the last part is not equal to end message
                        if message.endswith("endmessage"):
                            goon = False
                            break
                        #reset count, nonchar, and next char
                        count=1
                        nextcharacter=""
                    #if not enough pixels, add one to count, get the next value
                    count+=1
                    old = px[i,j][k]
                    value = str(old%2)
                    nextcharacter+=value
            else: break
    else: break

print("\n- - - The concealed message was: {", message[:-10], "} - - -")


            
                    
                
                
            
    











