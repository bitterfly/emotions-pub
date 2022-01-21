import sys
import xml.etree.ElementTree as ET
import time
import datetime
import math

def get_audio(log):
    beep_end = 0
    last_audio_file = ""
    last_audio_file_start = 0
    last_audio_file_end = 0
    audio = []

    with open(log, "r") as l:
        for line in l:
            spl = line.strip().split(" ")
            if spl[0] == "start":
                beep_end = float(spl[2])
            elif spl[0] == "audio":
                if spl[1] != last_audio_file:
                    # start new audio file
                    if last_audio_file_start != 0:
                        print("%s %f %f" % (last_audio_file, last_audio_file_start - beep_end, last_audio_file_end - last_audio_file_start))
                    last_audio_file_start = float(spl[2])
                    last_audio_file = spl[1]
                last_audio_file_end = float(spl[-1])
        print("%s %f %f" % (last_audio_file, last_audio_file_start - beep_end, last_audio_file_end - last_audio_file_start))
 
def get_other(log, tag):
    beep_end = 0

    with open(log, "r") as l:
        for line in l:
            spl = line.strip().split(" ")
            if spl[0] == "start":
                beep_end = float(spl[2])
            elif spl[0] == tag:
                if tag == "text":
                    print("neutral_%02d.wav %f %f" % (int(spl[1]), float(spl[2]) - beep_end, float(spl[-1]) - float(spl[2])))
                else:         
                    print("%s %f %f" % (spl[1], float(spl[2]) - beep_end, float(spl[-1]) - float(spl[2])))

if __name__ == "__main__":
    log = sys.argv[1]
    tag = sys.argv[2]
    
    if tag == "audio":
        get_audio(log)
    else:
        get_other(log, tag)
