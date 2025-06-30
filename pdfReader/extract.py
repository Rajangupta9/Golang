# extract.py
import sys
from pdfminer.high_level import extract_text

if len(sys.argv) < 2:
    print("Usage: python extract.py <filename>")
    sys.exit(1)

filename = sys.argv[1]

try:
    text = extract_text(filename)
    print(text)
except Exception as e:
    print("Error:", e)
    sys.exit(1)
