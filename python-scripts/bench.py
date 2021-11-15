from PIL import Image
import string
import random

import stego
import samplepairs

IMAGE = "testfiles/samplesmaller.png"
OUTPUT = "testfiles/testout.png"
END = "endmessage"
TRIALS = 5

with Image.open(IMAGE) as im:
    width, height = im.size

    bits = width * height * 3
    letters = bits // 8
    letters_end = letters - len(END)

    print("Letters\tBits\t% of file stegoed\tStego probability")

    for l in range(10, letters_end, letters_end // 20):
        bits_total = bits * 8
        bits_used = l * 8
        message_percent = bits_used / bits_total

        for _ in range(TRIALS):
            # from https://stackoverflow.com/questions/2257441/random-string-generation-with-upper-case-letters-and-digits
            message = ''.join(random.choice(string.ascii_uppercase + string.digits) for _ in range(l))

            message += END

            # Convert the message to binary
            binstring = ''.join(format(ord(i), '08b') for i in message)

            if not stego.stegoImage(im, binstring, OUTPUT, quiet=True):
                print("Failed to stego image.")
                exit(1)

            with Image.open(OUTPUT) as out:
                _, probability = samplepairs.analyzeSamplePairs(out)
                print("\t".join([str(a) for a in [l, bits_used, message_percent, probability]]))
