import sys
import xml.etree.ElementTree as ET
import time
import datetime
import math

if __name__ == "__main__":
    file = sys.argv[1]
    start = float(sys.argv[2])
    end = float(sys.argv[3])
    output_file = sys.argv[4]

    tree = ET.parse(file)
    s_date = tree.find("StartRecordingDate").text[:-3]
    s_time = tree.find("StartRecordingTime").text
    eeg_start = time.mktime(datetime.datetime.strptime(s_date+" "+s_time, "%d.%m.%Y %H:%M:%S.%f").timetuple())

    start_tick = math.ceil((start - eeg_start) * 500)
    end_tick = math.floor((end - eeg_start) * 500)

    ticks = list(tree.findall(".//tick"))

    with open(output_file, "w") as of:
        for d in ticks[start_tick:end_tick]:
            of.write("%s\n" % d.text.strip().replace(",", "."))