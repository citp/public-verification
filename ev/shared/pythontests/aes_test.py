from Crypto.Cipher import AES
from random import getrandbits

KEYBYTES = 16 # bytes
ENDIAN = "big"

def aes_enc(msg_in, key_in):
    """Use AES as a block cipher"""
    if type(msg_in) == int:
        msg = msg_in.to_bytes(KEYBYTES, ENDIAN)
    else:
        msg = msg_in

    if type(key_in) == int:
        key = key_in.to_bytes(KEYBYTES, ENDIAN)
    else:
        key = key_in

    cipher = AES.new(key, AES.MODE_ECB)
    return int.from_bytes(cipher.encrypt(msg), ENDIAN)

def run_test(msg, key, expected, goal=True):
    actual = aes_enc(msg, key)
    assert(goal == (actual == expected))


#run_test(msg = 0xffffffffffffffff0000000000000001,
        #key = 0x0000000000000000ffffffffffffffff, 
        #expected = 0x406bab6335ce415f4f943dc8966682aa
        #)



# test vectors from https://nvlpubs.nist.gov/nistpubs/FIPS/NIST.FIPS.197.pdf
run_test(msg = 0x00112233445566778899aabbccddeeff, 
        key = 0x000102030405060708090a0b0c0d0e0f,
        expected = 0x69c4e0d86a7b0430d8cdb78070b4c55a
        )

run_test(msg = 0x00112233445566778899aabbccddeefe, # flipped one bit
        key = 0x000102030405060708090a0b0c0d0e0f,
        expected = 0x69c4e0d86a7b0430d8cdb78070b4c55a,
        goal = False
        )

#print("AES(0,0): ",hex(aes_enc(0x0, 0x0)).upper())
#print("AES(msg=0,key=-1): ",hex(aes_enc(0x0, 0xffffffffffffffffffffffffffffffff)).upper())
def pprint(n, pref=""):
    print(pref + str(hex(n))[2:].rjust(32, "0").upper())

# test stuff
##msg_in = 0xffffffffffffffff0000000000000001
##key_in = 0x0000000000000000ffffffffffffffff
##msg_in = 0x00000000ffffffff0000000000000001
##key_in = 0x000000000000000000000000ffffffff
#msg_in = 0x0000000000000001ffffffffffffffff
#key_in = 0xffffffffffffffff0000000000000000
ika = getrandbits(128)
adkey = getrandbits(128)
ikb = ika
ek = getrandbits(128)
nonce = getrandbits(128)
msg_in = adkey^nonce if ika==ikb else nonce
key_in = ek
out = aes_enc(msg_in, key_in)
pprint(ika, "ika: ")
pprint(adkey, "adkey: ")
pprint(ikb, "ikb: ")
pprint(ek, "ek: ")
pprint(nonce, "nonce: ")
pprint(out, "out: ")
#out1 = int.from_bytes(out.to_bytes(16, ENDIAN)[0:8], ENDIAN)
#out2 = int.from_bytes(out.to_bytes(16, ENDIAN)[8:16], ENDIAN)
#print(out1)
#print(out2)
#
## Should be 0x406bab6335ce415f4f943dc8966682aa as per SCALE-MAMBA Documentation
## Each half of ct should be:
#4641992283528249695
#5734276157275538090
