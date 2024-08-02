from flask import Flask, request, render_template
import requests, os
import logging

app = Flask(__name__)
logging.basicConfig(level=logging.INFO) 

# TOKEN = "SECRET_TOKEN_VALUE"
TOKEN = os.environ.get("TOKEN")


@app.route('/')
def index():
    backend_url = 'http://backend'
    headers = {'Authorization': f'Bearer {TOKEN}'}
    app.logger.info('Sending request to backend')
    response = requests.get(backend_url, headers=headers)
    greeting = response.text
    return render_template('index.html', greeting=greeting)

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)