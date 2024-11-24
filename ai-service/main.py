from flask import Flask, jsonify
from flask_cors import CORS

app = Flask(__name__)
CORS(app, resources={r"/api/*": {"origins": "http://localhost:7012"}})  # Allow core-service

@app.route('/api/data', methods=['GET'])
def get_data():
    return jsonify({"data": "AI Service Data"})

if __name__ == '__main__':
    app.run(host="0.0.0.0", port=7011)
