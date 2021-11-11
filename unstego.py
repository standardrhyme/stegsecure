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
nonchar = 0
#loop through pixels
for i in range(imageheight):
    if nonchar>5: break #if there are more than 5 non-character values in a row, break
    for j in range(imagewidth):
        if nonchar>5: break
        else:
            for k in range(0,3):
                #if we have enough pixels to create new character
                if count>8:
                    nextcharacterbinary = int(nextcharacter,2)
                    charactertoadd = chr(nextcharacterbinary)
                    #if the character is between 0 and z
                    if nextcharacterbinary>31 and nextcharacterbinary<128:
                        #add to guess
                        message+=charactertoadd
                        #reset count, nonchar, and next char
                        count=1
                        nonchar=0
                        nextcharacter=""
                    else:
                        nonchar+=1
                #if not enough pixels, add one to count, get the next value
                count+=1
                old = px[i,j][k]
                value = str(old%2)
                nextcharacter+=value


print("\n\nThe concealed message was:", message)


            
                    
                
                
            
    











