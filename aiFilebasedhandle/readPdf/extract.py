# extract_api.py

from flask import Flask, request, jsonify
from pdfminer.high_level import extract_text
import os

app = Flask(__name__)

@app.route('/extract-pdf', methods=['POST'])
def extract_pdf():
    if 'file' not in request.files:
        return jsonify({'error': 'No file part'}), 400
    
    file = request.files['file']
    if file.filename == '':
        return jsonify({'error': 'No selected file'}), 400

    try:
        file_path = os.path.join("/tmp", file.filename)
        file.save(file_path)

        # Extract text using pdfminer
        text = extract_text(file_path)

        # Clean up temp file
        os.remove(file_path)

        return jsonify({'text': text})
    except Exception as e:
        return jsonify({'error': str(e)}), 500

if __name__ == '__main__':
    app.run(debug=True, port=5001)
