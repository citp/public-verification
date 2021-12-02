import re
import sys

# handle first three lines
# gates wires
# inA inB out

# remaining wires
# ins outs in1 in2opt out opcode

p_line1 = re.compile("(\d+)\s+(\d+)")
p_line2 = re.compile("(\d+)\s+(\d+)\s+(\d+)")
# line 3 blank

p_getnumop = re.compile("(\d+).*")
p_gate2in = re.compile("(\d)\s+(\d)\s+(\d+)\s+(\d+)\s+(\d+)\s+([a-zA-Z]+)")
p_gate1in = re.compile("(\d)\s+(\d)\s+(\d+)\s+(\d+)\s+([a-zA-Z]+)")

if len(sys.argv) < 3:
    print("Usage: python aes_to_ev_circuit.py AES-non-expanded.txt ev_notif_aes.txt")

with open(sys.argv[1], "r") as aes_circ:
    with open(sys.argv[2], "w") as ev_circ:

        # handle line 1
        line1 = aes_circ.readline()
        m_line1 = p_line1.match(line1)
        n_gates_orig, n_wires_orig = m_line1.groups()
        n_gates_new = str(int(n_gates_orig) + 641)
        n_wires_new = str(int(n_wires_orig) + 1025)
        ev_circ.write(f"{n_gates_new} {n_wires_new}\n")

        # handle line 2
        line2 = aes_circ.readline()
        m_line2 = p_line2.match(line2)
        n_in1_old, n_in2_old, n_out = m_line2.groups()
        n_in1_new, n_in2_new = "256", "384"
        ev_circ.write(f"{n_in1_new} {n_in2_new}   {n_out}\n")

        # handle line 3
        aes_circ.readline()
        ev_circ.write("\n")

        # write in new input gates
        # 0-255: CLIENT INPUT
        #   0-127: ika
        #   128-255: adkey
        # 256-639: SERVER INPUT
        #   256-383: ikb
        #   384-511: ek
        #   512-639: nonce (in real protocol, hardcode this.  used as input for convenience)
        # 640-767: A = ika XOR ikb (each cell 0 if equal, 1 ow)
        # 768: INV wire 0
        # 769: 1 = XOR 0 768
        # 770-897: B = A XOR fffff (each cell 1 if equal, 0 ow)
        # 898-1024: C = AND tree where 1022 is the AND of all elems of B
        # 1025-1152: D = CCCC AND adkey (either adkey or all 0s)
        # 1153-1280: E = D XOR nonce (the message to encrypt)
        for i in range(640, 768):
            ev_circ.write(f"2 1 {i-640} {i-640+256} {i} XOR\n")

        ev_circ.write(f"1 1 0 768 INV\n")
        ev_circ.write(f"2 1 0 768 769 XOR\n") # 769 = hardcoded 1

        for i in range(770, 898):
            ev_circ.write(f"2 1 769 {i-130} {i} XOR\n")

        for i in range(898, 962):
            ev_circ.write(f"2 1 {i-128} {i-127} {i} AND\n")
        for i in range(962, 994):
            ev_circ.write(f"2 1 {i-64} {i-63} {i} AND\n")
        for i in range(994, 1010):
            ev_circ.write(f"2 1 {i-32} {i-31} {i} AND\n")
        for i in range(1010, 1018):
            ev_circ.write(f"2 1 {i-16} {i-16} {i} AND\n")
        for i in range(1018, 1022):
            ev_circ.write(f"2 1 {i-8} {i-7} {i} AND\n")
        for i in range(1022, 1024):
            ev_circ.write(f"2 1 {i-4} {i-3} {i} AND\n")
        for i in range(1024, 1025):
            ev_circ.write(f"2 1 {i-2} {i-1} {i} AND\n") # 1024 = 111...1 if ika = ikb, 000...0 otherwise.

        for i in range(1025, 1153):
            ev_circ.write(f"2 1 1024 {i-1025+128} {i} AND\n")

        for i in range(1153, 1281):
            ev_circ.write(f"2 1 {i-128} {i-1153+512} {i} XOR\n") # outputs: 1153-1281 is either nonce, or nonceXORadkey



        # add aes gates
        # must change all instances of 0-127 to 1153-1280, and
        # also change all instances of 128-255 to 384-511
        # and all other wires +1025 (= 1281-256)
        def int_wirechanger(w):
            if 0 <= w <= 127:
                return w + 1153 # input message should be E
            elif 128 <= w <= 255:
                return w + 384 - 128 # input key should be ek
            else:
                return w + 1281 - 256 # all other wires have their id increased by 1025
        def wirechanger(w):
            return str(int_wirechanger(int(w)))
        for g in range(int(n_gates_orig)):
            line = aes_circ.readline()
            m_numop = p_getnumop.match(line)
            numop = int(m_numop.groups()[0])
            if numop == 1:
                m_line = p_gate1in.match(line)
                n_in, n_out, in1, out1, opcode = m_line.groups()
                ev_circ.write(f"{n_in} {n_out} {wirechanger(in1)} {wirechanger(out1)} {opcode}\n")
            elif numop == 2:
                m_line = p_gate2in.match(line)
                n_in, n_out, in1, in2, out1, opcode = m_line.groups()
                ev_circ.write(f"{n_in} {n_out} {wirechanger(in1)} {wirechanger(in2)} {wirechanger(out1)} {opcode}\n")
            else:
                print("Error")
                exit(0)
        print("Done!")

