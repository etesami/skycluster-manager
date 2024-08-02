from flask import Flask, request
import requests
import logging

app = Flask(__name__)
logging.basicConfig(level=logging.INFO) 

def verify_token(token):
    # Here you can implement your token verification logic
    # For simplicity, let's assume a hardcoded token for demonstration
    valid_token = "SECRET_TOKEN_VALUE"
    return token == valid_token

@app.route('/')
def hello():
    bearer_token = request.headers.get('Authorization')
    if bearer_token:
        # Check if the token starts with 'Bearer ' and extract the token value
        token = bearer_token.split('Bearer ')[-1]
        if verify_token(token):
            app.logger.info('Token verified')
            payment_response = requests.post('http://payment', json={'amount': 120})
            if payment_response.status_code == 200:
                app.logger.info('Payment processed successfully')
                # return 'Hello from the Backend! Payment processed successfully.'
                return 'Hello from the Backend! Payment processed successfully.' + payment_response.json().get('message')
            else:
                app.logger.info('Payment processing failed')
                return 'Payment processing failed.', 500
            # return 'Helloooo From the Server'
        else:
            return 'Unauthorized', 401
    else:
        return 'Unauthorized', 401

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
